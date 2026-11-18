function NavButton({ active, onClick, children }) {
  return (
    <button
      onClick={onClick}
      className={`text-left px-3 py-2 rounded-lg text-sm font-medium transition ${
        active ? 'bg-emerald-600 text-white' : 'hover:bg-gray-800 text-gray-400'
      }`}
    >
      {children}
    </button>
  );
}

// Left column: session info and navigation between the tabs.
export default function Sidebar({ isAdmin, activeTab, onSelect, enrollmentCount }) {
  return (
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
          <NavButton active={activeTab === 'courses'} onClick={() => onSelect('courses')}>
            Cursos Disponíveis
          </NavButton>
          {isAdmin && (
            <NavButton active={activeTab === 'createCourse'} onClick={() => onSelect('createCourse')}>
              Novo Curso
            </NavButton>
          )}
          <NavButton active={activeTab === 'enrollments'} onClick={() => onSelect('enrollments')}>
            Matrículas ({enrollmentCount})
          </NavButton>
        </div>
      </div>
    </div>
  );
}
