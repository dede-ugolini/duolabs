// "Matrículas" tab. A student only sees (and can cancel) their own
// enrollments; the admin sees everyone's.
export default function EnrollmentList({ enrollments, courses, isAdmin, onRefresh, onUnenroll }) {
  const courseTitle = (courseId) =>
    courses.find((c) => c.ID === courseId)?.title ?? 'Curso removido';

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <h2 className="text-xl font-bold text-gray-100">Matrículas Ativas</h2>
        <button
          onClick={onRefresh}
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
                <p className="text-sm font-semibold text-white">{courseTitle(en.course_id)}</p>
                <p className="text-xs text-gray-400">
                  Matrícula ID: {en.ID} | Curso ID: {en.course_id}
                  {isAdmin && ` | Usuário ID: ${en.user_id}`}
                </p>
              </div>
              <button
                onClick={() => onUnenroll(en.ID)}
                className="bg-red-950/40 hover:bg-red-900/50 text-red-400 border border-red-900/60 text-xs px-3 py-1.5 rounded-lg transition"
              >
                Cancelar Matrícula
              </button>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
