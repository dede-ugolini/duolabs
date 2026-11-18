import { useState, useEffect } from 'react';
import * as api from './api';
import { claimsFromToken, loadToken, saveToken, clearToken } from './auth';
import AuthForm from './components/AuthForm';
import Header from './components/Header';
import Sidebar from './components/Sidebar';
import CourseList from './components/CourseList';
import EnrollmentList from './components/EnrollmentList';
import CoursePlayer from './components/CoursePlayer';
import CourseEditor from './components/CourseEditor';

const isNotAllowed = (err) => err.status === 401 || err.status === 403;

export default function App() {
  const [token, setToken] = useState(loadToken);
  const [courses, setCourses] = useState([]);
  const [enrollments, setEnrollments] = useState([]);
  const [activeTab, setActiveTab] = useState('courses'); // 'courses', 'enrollments', 'createCourse', 'play'
  const [playingCourse, setPlayingCourse] = useState(null);

  // The claims only adapt the UI; the server is what enforces permissions.
  const claims = claimsFromToken(token);
  const isAdmin = claims.role === 'admin';
  const myEnrollments = enrollments.filter((en) => en.user_id === claims.user_id);
  const enrolledCourseIds = new Set(myEnrollments.map((en) => en.course_id));
  const visibleEnrollments = isAdmin ? enrollments : myEnrollments;

  const refreshCourses = async () => {
    try {
      const data = await api.listCourses();
      setCourses(Array.isArray(data) ? data : []);
    } catch (err) {
      console.error('Erro ao buscar cursos:', err);
    }
  };

  const refreshEnrollments = async () => {
    try {
      const data = await api.listEnrollments();
      setEnrollments(Array.isArray(data) ? data : []);
    } catch (err) {
      console.error('Erro ao buscar matrículas:', err);
    }
  };

  useEffect(() => {
    refreshCourses();
    refreshEnrollments();
  }, []);

  const handleLogin = (newToken) => {
    saveToken(newToken);
    setToken(newToken);
  };

  const handleLogout = () => {
    clearToken();
    setToken('');
    setActiveTab('courses');
    setPlayingCourse(null);
  };

  // Returns true when the course was created, so the editor can clear itself.
  const handleCreateCourse = async (payload) => {
    try {
      await api.createCourse(token, payload);
      alert('Curso criado com sucesso!');
      setActiveTab('courses');
      refreshCourses();
      return true;
    } catch (err) {
      alert(isNotAllowed(err) ? 'Apenas administradores podem criar cursos' : err.message);
      return false;
    }
  };

  const handleDeleteCourse = async (id) => {
    if (!confirm('Deseja realmente deletar este curso?')) return;
    try {
      await api.deleteCourse(token, id);
      refreshCourses();
    } catch (err) {
      alert(isNotAllowed(err) ? 'Apenas administradores podem deletar cursos' : err.message);
    }
  };

  const handleEnroll = async (courseId) => {
    try {
      await api.enroll(token, courseId);
      alert('Matrícula realizada com sucesso!');
      refreshEnrollments();
    } catch (err) {
      alert(err.status === 409 ? 'Você já está matriculado neste curso.' : err.message);
    }
  };

  const handleUnenroll = async (enrollmentId) => {
    try {
      await api.unenroll(token, enrollmentId);
      alert('Matrícula cancelada!');
      refreshEnrollments();
    } catch (err) {
      alert(err.message);
    }
  };

  const handlePlay = (course) => {
    setPlayingCourse(course);
    setActiveTab('play');
  };

  if (!token) return <AuthForm onLogin={handleLogin} />;

  return (
    <div className="min-h-screen bg-gray-950 text-gray-100 flex flex-col font-sans">
      <Header isAdmin={isAdmin} onLogout={handleLogout} />

      <div className="flex flex-1 max-w-7xl w-full mx-auto p-6 gap-6">
        <Sidebar
          isAdmin={isAdmin}
          activeTab={activeTab}
          onSelect={setActiveTab}
          enrollmentCount={visibleEnrollments.length}
        />

        <div className="flex-1 min-w-0 bg-gray-900 border border-gray-800 rounded-2xl p-6 shadow-xl">
          {activeTab === 'courses' && (
            <CourseList
              courses={courses}
              enrolledCourseIds={enrolledCourseIds}
              isAdmin={isAdmin}
              onRefresh={refreshCourses}
              onPlay={handlePlay}
              onEnroll={handleEnroll}
              onDelete={handleDeleteCourse}
            />
          )}

          {activeTab === 'createCourse' && isAdmin && <CourseEditor onSubmit={handleCreateCourse} />}

          {activeTab === 'play' && playingCourse && (
            <CoursePlayer
              key={playingCourse.ID}
              token={token}
              course={playingCourse}
              onClose={() => setActiveTab('courses')}
              onProgressChange={refreshEnrollments}
            />
          )}

          {activeTab === 'enrollments' && (
            <EnrollmentList
              enrollments={visibleEnrollments}
              courses={courses}
              isAdmin={isAdmin}
              onRefresh={refreshEnrollments}
              onUnenroll={handleUnenroll}
            />
          )}
        </div>
      </div>
    </div>
  );
}
