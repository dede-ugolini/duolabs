import { useState } from 'react';
import { login, register } from '../api';

const fieldClass =
  'w-full bg-gray-950 border border-gray-800 rounded-xl px-4 py-3 text-sm focus:outline-none focus:border-emerald-500 transition-colors';
const labelClass = 'block text-xs font-semibold text-gray-400 uppercase tracking-wider mb-2';

// Login / sign-up screen. Calls onLogin(token) after a successful login.
export default function AuthForm({ onLogin }) {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [mode, setMode] = useState('login'); // 'login' or 'register'
  const isLogin = mode === 'login';

  const handleSubmit = async (e) => {
    e.preventDefault();
    try {
      if (isLogin) {
        const { token } = await login(username, password);
        alert('Login realizado com sucesso!');
        onLogin(token);
      } else {
        await register(username, password);
        alert('Cadastro realizado com sucesso! Faça login.');
        setMode('login');
      }
    } catch (err) {
      alert(err.message || 'Erro na autenticação');
    }
  };

  return (
    <div className="min-h-screen bg-gray-950 text-gray-100 flex flex-col font-sans items-center justify-center p-6">
      <div className="max-w-md w-full bg-gray-900 border border-gray-800 rounded-2xl p-8 shadow-2xl">
        <div className="flex flex-col items-center mb-8">
          <span className="text-4xl font-black bg-gradient-to-r from-green-400 to-emerald-500 bg-clip-text text-transparent mb-2">
            Duolabs
          </span>
          <p className="text-gray-400 text-sm text-center">
            {isLogin
              ? 'Faça login para acessar os laboratórios.'
              : 'Crie sua conta para começar a aprender.'}
          </p>
        </div>

        <form onSubmit={handleSubmit} className="space-y-5">
          <div>
            <label className={labelClass}>Usuário</label>
            <input
              type="text"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              required
              className={fieldClass}
              placeholder="Seu usuário"
            />
          </div>
          <div>
            <label className={labelClass}>Senha</label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              className={fieldClass}
              placeholder="••••••••"
            />
          </div>
          <button
            type="submit"
            className="w-full bg-emerald-600 hover:bg-emerald-500 text-white font-bold py-3 rounded-xl text-sm transition-all transform hover:scale-[1.02] active:scale-[0.98] shadow-lg shadow-emerald-900/20"
          >
            {isLogin ? 'Entrar Agora' : 'Criar Conta'}
          </button>
        </form>

        <div className="mt-8 pt-6 border-t border-gray-800 text-center">
          <button
            onClick={() => setMode(isLogin ? 'register' : 'login')}
            className="text-xs text-emerald-400 hover:text-emerald-300 font-medium transition-colors"
          >
            {isLogin ? 'Não tem uma conta? Cadastre-se' : 'Já possui uma conta? Acesse aqui'}
          </button>
        </div>
      </div>
    </div>
  );
}
