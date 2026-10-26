import React, { useState, useEffect } from 'react';

const API_BASE = 'http://localhost:8081';

export default function App() {
  const [token, setToken] = useState(localStorage.getItem('duolabs_token') || '');
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [authMode, setAuthMode] = useState('login'); // 'login' or 'register'
  
  const [courses, setCourses] = useState([]);
  const [enrollments, setEnrollments] = useState([]);
  const [activeTab, setActiveTab] = useState('courses'); // 'courses', 'enrollments', 'createCourse'

  // New Course Form state
  const [newTitle, setNewTitle] = useState('');
  const [newDesc, setNewDesc] = useState('');
  const [newCategory, setNewCategory] = useState('');
  const [newDifficulty, setNewDifficulty] = useState('beginner');
  const [newXp, setNewXp] = useState(500);

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
      const data = await res.json();
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

  const handleCreateCourse = async (e) => {
    e.preventDefault();
    try {
      const payload = {
        title: newTitle,
        description: newDesc,
        category: newCategory,
        difficulty: newDifficulty,
        xp: parseInt(newXp),
        units: []
      };
      const res = await fetch(`${API_BASE}/courses`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload)
      });
      if (!res.ok) throw new Error('Erro ao criar curso');
      alert('Curso criado com sucesso!');
      setNewTitle('');
      setNewDesc('');
      setNewCategory('');
      setActiveTab('courses');
      fetchCourses();
    } catch (err) {
      alert(err.message);
    }
  };

  const handleDeleteCourse = async (id) => {
    if (!confirm('Deseja realmente deletar este curso?')) return;
    try {
      const res = await fetch(`${API_BASE}/courses/${id}`, { method: 'DELETE' });
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
          {token ? (
            <div className="flex items-center space-x-3">
              <span className="text-sm text-gray-300">Autenticado 🟢</span>
              <button
                onClick={handleLogout}
                className="bg-red-600/20 hover:bg-red-600/30 text-red-400 border border-red-800 text-xs px-3 py-1.5 rounded-lg font-semibold transition"
              >
                Sair
              </button>
            </div>
          ) : (
            <span className="text-xs text-yellow-500 bg-yellow-950/50 border border-yellow-800/50 px-2.5 py-1 rounded-md">
              Não autenticado (Faça login para matricular)
            </span>
          )}
        </div>
      </header>

      {/* Main Content */}
      <div className="flex flex-1 max-w-7xl w-full mx-auto p-6 gap-6">
        {/* Sidebar / Auth panel */}
        <div className="w-80 flex flex-col gap-6">
          {!token ? (
            <div className="bg-gray-900 border border-gray-800 rounded-2xl p-5 shadow-xl">
              <h2 className="text-lg font-bold mb-4 text-gray-200">
                {authMode === 'login' ? '🔑 Entrar' : '📝 Criar Conta'}
              </h2>
              <form onSubmit={handleAuth} className="space-y-4">
                <div>
                  <label className="block text-xs text-gray-400 mb-1">Usuário</label>
                  <input
                    type="text"
                    value={username}
                    onChange={(e) => setUsername(e.target.value)}
                    required
                    className="w-full bg-gray-950 border border-gray-800 rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-emerald-500"
                    placeholder="ex: student1"
                  />
                </div>
                <div>
                  <label className="block text-xs text-gray-400 mb-1">Senha</label>
                  <input
                    type="password"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    required
                    className="w-full bg-gray-950 border border-gray-800 rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-emerald-500"
                    placeholder="••••••"
                  />
                </div>
                <button
                  type="submit"
                  className="w-full bg-emerald-600 hover:bg-emerald-500 text-white font-semibold py-2 rounded-lg text-sm transition shadow-lg shadow-emerald-900/20"
                >
                  {authMode === 'login' ? 'Entrar' : 'Cadastrar'}
                </button>
              </form>
              <button
                onClick={() => setAuthMode(authMode === 'login' ? 'register' : 'login')}
                className="mt-4 w-full text-xs text-emerald-400 hover:underline text-center block"
              >
                {authMode === 'login' ? 'Não tem conta? Cadastre-se' : 'Já tem conta? Faça login'}
              </button>
            </div>
          ) : (
            <div className="bg-emerald-950/20 border border-emerald-900/40 rounded-2xl p-5 shadow-xl">
              <h2 className="text-sm font-bold text-emerald-400 mb-2">Sessão Ativa</h2>
              <p className="text-xs text-gray-300">Você está logado e pronto para gerenciar matrículas.</p>
            </div>
          )}

          <div className="bg-gray-900 border border-gray-800 rounded-2xl p-5 shadow-xl">
            <h3 className="text-sm font-bold mb-3 text-gray-300">Navegação</h3>
            <div className="flex flex-col gap-2">
              <button
                onClick={() => setActiveTab('courses')}
                className={`text-left px-3 py-2 rounded-lg text-sm font-medium transition ${activeTab === 'courses' ? 'bg-emerald-600 text-white' : 'hover:bg-gray-800 text-gray-400'}`}
              >
                📚 Cursos Disponíveis
              </button>
              <button
                onClick={() => setActiveTab('createCourse')}
                className={`text-left px-3 py-2 rounded-lg text-sm font-medium transition ${activeTab === 'createCourse' ? 'bg-emerald-600 text-white' : 'hover:bg-gray-800 text-gray-400'}`}
              >
                ➕ Novo Curso
              </button>
              <button
                onClick={() => setActiveTab('enrollments')}
                className={`text-left px-3 py-2 rounded-lg text-sm font-medium transition ${activeTab === 'enrollments' ? 'bg-emerald-600 text-white' : 'hover:bg-gray-800 text-gray-400'}`}
              >
                🎓 Matrículas ({enrollments.length})
              </button>
            </div>
          </div>
        </div>

        {/* Content Area */}
        <div className="flex-1 bg-gray-900 border border-gray-800 rounded-2xl p-6 shadow-xl">
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
                          <button
                            onClick={() => handleEnroll(course.ID)}
                            className="bg-emerald-600/20 hover:bg-emerald-600/30 text-emerald-400 border border-emerald-800 text-xs px-3 py-1.5 rounded-lg font-semibold transition"
                          >
                            Matricular
                          </button>
                          <button
                            onClick={() => handleDeleteCourse(course.ID)}
                            className="bg-red-950/40 hover:bg-red-900/50 text-red-400 border border-red-900/60 text-xs px-3 py-1.5 rounded-lg transition"
                          >
                            Deletar
                          </button>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          )}

          {activeTab === 'createCourse' && (
            <div>
              <h2 className="text-xl font-bold text-gray-100 mb-6">Criar Novo Curso</h2>
              <form onSubmit={handleCreateCourse} className="space-y-4 max-w-lg">
                <div>
                  <label className="block text-xs text-gray-400 mb-1">Título</label>
                  <input
                    type="text"
                    value={newTitle}
                    onChange={(e) => setNewTitle(e.target.value)}
                    required
                    className="w-full bg-gray-950 border border-gray-800 rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-emerald-500"
                    placeholder="Ex: Rust para Iniciantes"
                  />
                </div>
                <div>
                  <label className="block text-xs text-gray-400 mb-1">Descrição</label>
                  <textarea
                    value={newDesc}
                    onChange={(e) => setNewDesc(e.target.value)}
                    required
                    rows="3"
                    className="w-full bg-gray-950 border border-gray-800 rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-emerald-500"
                    placeholder="Resumo do curso..."
                  />
                </div>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="block text-xs text-gray-400 mb-1">Categoria</label>
                    <input
                      type="text"
                      value={newCategory}
                      onChange={(e) => setNewCategory(e.target.value)}
                      required
                      className="w-full bg-gray-950 border border-gray-800 rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-emerald-500"
                      placeholder="Programação"
                    />
                  </div>
                  <div>
                    <label className="block text-xs text-gray-400 mb-1">Dificuldade</label>
                    <select
                      value={newDifficulty}
                      onChange={(e) => setNewDifficulty(e.target.value)}
                      className="w-full bg-gray-950 border border-gray-800 rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-emerald-500"
                    >
                      <option value="beginner">Beginner</option>
                      <option value="intermediate">Intermediate</option>
                      <option value="advanced">Advanced</option>
                    </select>
                  </div>
                </div>
                <div>
                  <label className="block text-xs text-gray-400 mb-1">XP</label>
                  <input
                    type="number"
                    value={newXp}
                    onChange={(e) => setNewXp(e.target.value)}
                    required
                    className="w-full bg-gray-950 border border-gray-800 rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-emerald-500"
                  />
                </div>
                <button
                  type="submit"
                  className="bg-emerald-600 hover:bg-emerald-500 text-white font-semibold px-6 py-2 rounded-lg text-sm transition shadow-lg shadow-emerald-900/20"
                >
                  Salvar Curso
                </button>
              </form>
            </div>
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
