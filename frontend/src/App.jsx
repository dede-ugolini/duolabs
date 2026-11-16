import React, { useState, useEffect } from 'react';
import CoursePlayer from './components/CoursePlayer';
import CourseEditor from './components/CourseEditor';

// In dev the API runs separately on :8081; the production build is served by
// the Go server itself, so it uses same-origin requests. Override with
// VITE_API_BASE when the API lives elsewhere.
const API_BASE =
  import.meta.env.VITE_API_BASE ??
  (import.meta.env.DEV ? 'http://localhost:8081' : '');

// Reads the claims from the JWT payload. Only used to adapt the UI (admin
// buttons, which courses the user is enrolled in); the server enforces
// the permissions.
function claimsFromToken(token) {
  try {
    const payload = token.split('.')[1].replace(/-/g, '+').replace(/_/g, '/');
    return JSON.parse(atob(payload));
  } catch {
    return {};
  }
}

export default function App() {
  const [token, setToken] = useState(localStorage.getItem('duolabs_token') || '');
  const claims = claimsFromToken(token);
  const isAdmin = claims.role === 'admin';
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [authMode, setAuthMode] = useState('login'); // 'login' or 'register'
  
  const [courses, setCourses] = useState([]);
  const [enrollments, setEnrollments] = useState([]);
  const [activeTab, setActiveTab] = useState('courses'); // 'courses', 'enrollments', 'createCourse', 'play'
  const [playingCourse, setPlayingCourse] = useState(null);

  const enrolledCourseIds = new Set(
    enrollments.filter((en) => en.user_id === claims.user_id).map((en) => en.course_id)
  );

  useEffect(() => {
    fetchCourses();
    fetchEnrollments();
  }, []);

  const handleAuth = async (e) => {
    e.preventDefault();
    const endpoint = authMode === 'login' ? '/login' : '/register';
    try {
      const res = await fetch(`${API_BASE}${endpoint}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name: username, password })
      });
      const raw = await res.text();
      let data = {};
      try { data = JSON.parse(raw); } catch { data = { error: raw.trim() }; }
      if (!res.ok) throw new Error(data.error || 'Erro na autenticação');
      
      if (authMode === 'login') {
        setToken(data.token);
        localStorage.setItem('duolabs_token', data.token);
        alert('Login realizado com sucesso!');
      } else {
        alert('Cadastro realizado com sucesso! Faça login.');
        setAuthMode('login');
      }
    } catch (err) {
      alert(err.message);
    }
  };

  const handleLogout = () => {
    setToken('');
    setActiveTab('courses');
    setPlayingCourse(null);
    localStorage.removeItem('duolabs_token');
  };

  const fetchCourses = async () => {
    try {
      const res = await fetch(`${API_BASE}/courses`);
      const data = await res.json();
      setCourses(Array.isArray(data) ? data : []);
    } catch (err) {
      console.error('Erro ao buscar cursos:', err);
    }
  };

  const fetchEnrollments = async () => {
    try {
      const res = await fetch(`${API_BASE}/enrollments`);
      const data = await res.json();
      setEnrollments(Array.isArray(data) ? data : []);
    } catch (err) {
      console.error('Erro ao buscar matrículas:', err);
    }
  };

  const handleCreateCourse = async (payload) => {
    try {
      const res = await fetch(`${API_BASE}/courses`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`
        },
        body: JSON.stringify(payload)
      });
      if (res.status === 401 || res.status === 403) throw new Error('Apenas administradores podem criar cursos');
      if (!res.ok) throw new Error((await res.text()).trim() || 'Erro ao criar curso');
      alert('Curso criado com sucesso!');
      setActiveTab('courses');
      fetchCourses();
      return true;
    } catch (err) {
      alert(err.message);
      return false;
    }
  };

  const handlePlay = (course) => {
    setPlayingCourse(course);
    setActiveTab('play');
  };

  const handleDeleteCourse = async (id) => {
    if (!confirm('Deseja realmente deletar este curso?')) return;
    try {
      const res = await fetch(`${API_BASE}/courses/${id}`, {
        method: 'DELETE',
        headers: { 'Authorization': `Bearer ${token}` }
      });
      if (res.status === 401 || res.status === 403) throw new Error('Apenas administradores podem deletar cursos');
      if (!res.ok) throw new Error('Erro ao deletar curso');
      fetchCourses();
    } catch (err) {
      alert(err.message);
    }
  };

  const handleEnroll = async (courseId) => {
    if (!token) {
      alert('Faça login para se matricular!');
      return;
    }
    try {
      const res = await fetch(`${API_BASE}/enrollments`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`
        },
        body: JSON.stringify({ course_id: courseId })
      });
      const data = await res.json();
      if (!res.ok) throw new Error(data.error || 'Erro ao matricular');
      alert('Matrícula realizada com sucesso!');
      fetchEnrollments();
    } catch (err) {
      alert(err.message);
    }
  };

  const handleUnenroll = async (enrollmentId) => {
    if (!token) {
      alert('Faça login para cancelar matrícula!');
      return;
    }
    try {
      const res = await fetch(`${API_BASE}/enrollments/${enrollmentId}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${token}`
        }
      });
      if (!res.ok) throw new Error('Erro ao cancelar matrícula');
      alert('Matrícula cancelada!');
      fetchEnrollments();
    } catch (err) {
      alert(err.message);
    }
  };

  if (!token) {
    return (
      <div className="min-h-screen bg-gray-950 text-gray-100 flex flex-col font-sans items-center justify-center p-6">
        <div className="max-w-md w-full bg-gray-900 border border-gray-800 rounded-2xl p-8 shadow-2xl">
          <div className="flex flex-col items-center mb-8">
            <span className="text-4xl font-black bg-gradient-to-r from-green-400 to-emerald-500 bg-clip-text text-transparent mb-2">
              Duolabs
            </span>
            <p className="text-gray-400 text-sm text-center">
              {authMode === 'login' 
                ? 'Faça login para acessar os laboratórios.' 
                : 'Crie sua conta para começar a aprender.'}
            </p>
          </div>

          <form onSubmit={handleAuth} className="space-y-5">
            <div>
              <label className="block text-xs font-semibold text-gray-400 uppercase tracking-wider mb-2">Usuário</label>
              <input
                type="text"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                required
                className="w-full bg-gray-950 border border-gray-800 rounded-xl px-4 py-3 text-sm focus:outline-none focus:border-emerald-500 transition-colors"
                placeholder="Seu usuário"
              />
            </div>
            <div>
              <label className="block text-xs font-semibold text-gray-400 uppercase tracking-wider mb-2">Senha</label>
              <input
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
                className="w-full bg-gray-950 border border-gray-800 rounded-xl px-4 py-3 text-sm focus:outline-none focus:border-emerald-500 transition-colors"
                placeholder="••••••••"
              />
            </div>
            <button
              type="submit"
              className="w-full bg-emerald-600 hover:bg-emerald-500 text-white font-bold py-3 rounded-xl text-sm transition-all transform hover:scale-[1.02] active:scale-[0.98] shadow-lg shadow-emerald-900/20"
            >
              {authMode === 'login' ? 'Entrar Agora' : 'Criar Conta'}
            </button>
          </form>

          <div className="mt-8 pt-6 border-t border-gray-800 text-center">
            <button
              onClick={() => setAuthMode(authMode === 'login' ? 'register' : 'login')}
              className="text-xs text-emerald-400 hover:text-emerald-300 font-medium transition-colors"
            >
              {authMode === 'login' ? 'Não tem uma conta? Cadastre-se' : 'Já possui uma conta? Acesse aqui'}
            </button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-950 text-gray-100 flex flex-col font-sans">
      {/* Header */}
      <header className="bg-gray-900 border-b border-gray-800 px-6 py-4 flex justify-between items-center shadow-md">
        <div className="flex items-center space-x-3">
          <span className="text-2xl font-black bg-gradient-to-r from-green-400 to-emerald-500 bg-clip-text text-transparent">
            Duolabs
          </span>
          <span className="text-xs bg-emerald-950 text-emerald-400 border border-emerald-800 px-2 py-0.5 rounded-full font-medium">
            Painel de Exemplo
          </span>
        </div>
        <div className="flex items-center space-x-4">
          <div className="flex items-center space-x-3">
            <span className="text-sm text-gray-300">{isAdmin ? 'Administrador 🛡️' : 'Autenticado 🟢'}</span>
            <button
              onClick={handleLogout}
              className="bg-red-600/20 hover:bg-red-600/30 text-red-400 border border-red-800 text-xs px-3 py-1.5 rounded-lg font-semibold transition"
            >
              Sair
            </button>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <div className="flex flex-1 max-w-7xl w-full mx-auto p-6 gap-6">
        {/* Sidebar */}
        <div className="w-80 flex flex-col gap-6">
          <div className="bg-emerald-950/20 border border-emerald-900/40 rounded-2xl p-5 shadow-xl">
            <h2 className="text-sm font-bold text-emerald-400 mb-2">Sessão Ativa</h2>
            <p className="text-xs text-gray-300">
              {isAdmin
                ? 'Você é administrador: pode criar e remover cursos.'
                : 'Você está logado e pronto para fazer cursos.'}
            </p>
          </div>
          <div className="bg-gray-900 border border-gray-800 rounded-2xl p-5 shadow-xl">
            <h3 className="text-sm font-bold mb-3 text-gray-300">Navegação</h3>
            <div className="flex flex-col gap-2">
              <button
                onClick={() => setActiveTab('courses')}
                className={`text-left px-3 py-2 rounded-lg text-sm font-medium transition ${activeTab === 'courses' ? 'bg-emerald-600 text-white' : 'hover:bg-gray-800 text-gray-400'}`}
              >
                 Cursos Disponíveis
              </button>
              {isAdmin && (
                <button
                  onClick={() => setActiveTab('createCourse')}
                  className={`text-left px-3 py-2 rounded-lg text-sm font-medium transition ${activeTab === 'createCourse' ? 'bg-emerald-600 text-white' : 'hover:bg-gray-800 text-gray-400'}`}
                >
                   Novo Curso
                </button>
              )}
              <button
                onClick={() => setActiveTab('enrollments')}
                className={`text-left px-3 py-2 rounded-lg text-sm font-medium transition ${activeTab === 'enrollments' ? 'bg-emerald-600 text-white' : 'hover:bg-gray-800 text-gray-400'}`}
              >
                 Matrículas ({enrollments.length})
              </button>
            </div>
          </div>
        </div>

        {/* Content Area */}
        <div className="flex-1 min-w-0 bg-gray-900 border border-gray-800 rounded-2xl p-6 shadow-xl">
          {activeTab === 'courses' && (
            <div>
              <div className="flex justify-between items-center mb-6">
                <h2 className="text-xl font-bold text-gray-100">Cursos</h2>
                <button
                  onClick={fetchCourses}
                  className="bg-gray-800 hover:bg-gray-700 text-xs px-3 py-1.5 rounded-lg text-gray-300 transition"
                >
                  🔄 Atualizar
                </button>
              </div>

              {courses.length === 0 ? (
                <p className="text-gray-500 text-sm">Nenhum curso cadastrado ainda.</p>
              ) : (
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  {courses.map((course) => (
                    <div key={course.ID} className="bg-gray-950 border border-gray-800/80 rounded-xl p-5 flex flex-col justify-between hover:border-gray-700 transition">
                      <div>
                        <div className="flex justify-between items-start mb-2">
                          <span className="text-xs bg-emerald-950 text-emerald-400 border border-emerald-800 px-2 py-0.5 rounded uppercase font-bold tracking-wider">
                            {course.category}
                          </span>
                          <span className="text-xs text-yellow-500 font-semibold bg-yellow-950/40 border border-yellow-900/50 px-2 py-0.5 rounded">
                            ✨ {course.xp} XP
                          </span>
                        </div>
                        <h3 className="text-lg font-bold text-white mb-1">{course.title}</h3>
                        <p className="text-xs text-gray-400 mb-4 line-clamp-2">{course.description}</p>
                      </div>

                      <div className="flex items-center justify-between pt-3 border-t border-gray-900">
                        <span className="text-xs text-gray-500 capitalize">Dificuldade: {course.difficulty}</span>
                        <div className="flex gap-2">
                          {enrolledCourseIds.has(course.ID) ? (
                            <button
                              onClick={() => handlePlay(course)}
                              className="bg-emerald-600 hover:bg-emerald-500 text-white text-xs px-3 py-1.5 rounded-lg font-semibold transition"
                            >
                              Fazer curso
                            </button>
                          ) : (
                            <button
                              onClick={() => handleEnroll(course.ID)}
                              className="bg-emerald-600/20 hover:bg-emerald-600/30 text-emerald-400 border border-emerald-800 text-xs px-3 py-1.5 rounded-lg font-semibold transition"
                            >
                              Matricular
                            </button>
                          )}
                          {isAdmin && (
                            <button
                              onClick={() => handleDeleteCourse(course.ID)}
                              className="bg-red-950/40 hover:bg-red-900/50 text-red-400 border border-red-900/60 text-xs px-3 py-1.5 rounded-lg transition"
                            >
                              Deletar
                            </button>
                          )}
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          )}

          {activeTab === 'createCourse' && isAdmin && (
            <CourseEditor onSubmit={handleCreateCourse} />
          )}

          {activeTab === 'play' && playingCourse && (
            <CoursePlayer
              key={playingCourse.ID}
              apiBase={API_BASE}
              token={token}
              course={playingCourse}
              onClose={() => setActiveTab('courses')}
              onProgressChange={fetchEnrollments}
            />
          )}

          {activeTab === 'enrollments' && (
            <div>
              <div className="flex justify-between items-center mb-6">
                <h2 className="text-xl font-bold text-gray-100">Matrículas Ativas</h2>
                <button
                  onClick={fetchEnrollments}
                  className="bg-gray-800 hover:bg-gray-700 text-xs px-3 py-1.5 rounded-lg text-gray-300 transition"
                >
                  🔄 Atualizar
                </button>
              </div>

              {enrollments.length === 0 ? (
                <p className="text-gray-500 text-sm">Nenhuma matrícula encontrada.</p>
              ) : (
                <div className="space-y-3">
                  {enrollments.map((en) => (
                    <div key={en.ID} className="bg-gray-950 border border-gray-800/80 rounded-xl p-4 flex justify-between items-center">
                      <div>
                        <p className="text-sm font-semibold text-white">Matrícula ID: {en.ID}</p>
                        <p className="text-xs text-gray-400">Usuário ID: {en.user_id} | Curso ID: {en.course_id}</p>
                      </div>
                      <button
                        onClick={() => handleUnenroll(en.ID)}
                        className="bg-red-950/40 hover:bg-red-900/50 text-red-400 border border-red-900/60 text-xs px-3 py-1.5 rounded-lg transition"
                      >
                        Cancelar Matrícula
                      </button>
                    </div>
                  ))}
                </div>
              )}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
