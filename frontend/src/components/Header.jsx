export default function Header({ isAdmin, onLogout }) {
  return (
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
            onClick={onLogout}
            className="bg-red-600/20 hover:bg-red-600/30 text-red-400 border border-red-800 text-xs px-3 py-1.5 rounded-lg font-semibold transition"
          >
            Sair
          </button>
        </div>
      </div>
    </header>
  );
}
