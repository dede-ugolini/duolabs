function CourseCard({ course, enrolled, isAdmin, onPlay, onEnroll, onDelete }) {
  return (
    <div className="bg-gray-950 border border-gray-800/80 rounded-xl p-5 flex flex-col justify-between hover:border-gray-700 transition">
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
          {enrolled ? (
            <button
              onClick={() => onPlay(course)}
              className="bg-emerald-600 hover:bg-emerald-500 text-white text-xs px-3 py-1.5 rounded-lg font-semibold transition"
            >
              Fazer curso
            </button>
          ) : (
            <button
              onClick={() => onEnroll(course.ID)}
              className="bg-emerald-600/20 hover:bg-emerald-600/30 text-emerald-400 border border-emerald-800 text-xs px-3 py-1.5 rounded-lg font-semibold transition"
            >
              Matricular
            </button>
          )}
          {isAdmin && (
            <button
              onClick={() => onDelete(course.ID)}
              className="bg-red-950/40 hover:bg-red-900/50 text-red-400 border border-red-900/60 text-xs px-3 py-1.5 rounded-lg transition"
            >
              Deletar
            </button>
          )}
        </div>
      </div>
    </div>
  );
}

// "Cursos" tab: one card per course.
export default function CourseList({ courses, enrolledCourseIds, isAdmin, onRefresh, onPlay, onEnroll, onDelete }) {
  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <h2 className="text-xl font-bold text-gray-100">Cursos</h2>
        <button
          onClick={onRefresh}
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
            <CourseCard
              key={course.ID}
              course={course}
              enrolled={enrolledCourseIds.has(course.ID)}
              isAdmin={isAdmin}
              onPlay={onPlay}
              onEnroll={onEnroll}
              onDelete={onDelete}
            />
          ))}
        </div>
      )}
    </div>
  );
}
