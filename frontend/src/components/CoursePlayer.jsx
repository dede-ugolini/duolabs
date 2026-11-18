import React, { useState, useEffect, useMemo } from 'react';
import { getCourse, getProgress, submitAnswer } from '../api';

function flattenLessons(course) {
  const lessons = [];
  (course?.units || []).forEach((unit) => {
    (unit.tasks || []).forEach((task) => {
      (task.lessons || []).forEach((lesson) => {
        lessons.push({ ...lesson, unitTitle: unit.title, taskTitle: task.title });
      });
    });
  });
  return lessons;
}

function shuffle(items) {
  const arr = [...items];
  for (let i = arr.length - 1; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1));
    [arr[i], arr[j]] = [arr[j], arr[i]];
  }
  return arr;
}

const primaryButton =
  'bg-emerald-600 hover:bg-emerald-500 text-white font-semibold px-6 py-2 rounded-lg text-sm transition';

export default function CoursePlayer({ token, course, onClose, onProgressChange }) {
  const [lessons, setLessons] = useState([]);
  const [completed, setCompleted] = useState(() => new Set());
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState('');
  const [index, setIndex] = useState(0);
  const [done, setDone] = useState(false);
  const [selected, setSelected] = useState(null);
  const [submitting, setSubmitting] = useState(false);
  const [result, setResult] = useState(null); // null | 'correct' | 'wrong'
  const [message, setMessage] = useState('');

  useEffect(() => {
    let active = true;
    const load = async () => {
      setLoading(true);
      setLoadError('');
      try {
        const [courseData, progress] = await Promise.all([
          getCourse(course.ID),
          getProgress(token, course.ID)
        ]);
        const flat = flattenLessons(courseData);
        const doneIds = new Set(progress.completed_lesson_ids || []);
        const first = flat.findIndex((l) => !doneIds.has(l.ID));
        if (!active) return;
        setLessons(flat);
        setCompleted(doneIds);
        setIndex(first >= 0 ? first : 0);
        setDone(flat.length > 0 && first < 0);
      } catch (err) {
        if (active) setLoadError(err.message);
      } finally {
        if (active) setLoading(false);
      }
    };
    load();
    return () => {
      active = false;
    };
  }, [course.ID, token]);

  const lesson = lessons[index];
  const choices = useMemo(() => shuffle(lesson?.choices || []), [lesson?.ID]);
  const completedCount = lessons.filter((l) => completed.has(l.ID)).length;
  const percent = lessons.length > 0 ? (completedCount / lessons.length) * 100 : 0;
  const locked = result === 'correct' || submitting;

  const pickChoice = (choiceId) => {
    if (locked) return;
    setSelected(choiceId);
    setResult(null);
    setMessage('');
  };

  const submit = async () => {
    if (!selected || locked) return;
    setSubmitting(true);
    setMessage('');
    try {
      const data = await submitAnswer(token, lesson.ID, selected);
      if (data.correct) {
        setResult('correct');
        setCompleted((prev) => new Set(prev).add(lesson.ID));
        onProgressChange?.();
      } else {
        setResult('wrong');
      }
    } catch (err) {
      setMessage(err.status === 403 ? 'Matricule-se neste curso para responder.' : err.message);
    } finally {
      setSubmitting(false);
    }
  };

  const next = () => {
    const nextIndex = lessons.findIndex((l, i) => i > index && !completed.has(l.ID));
    setResult(null);
    setSelected(null);
    setMessage('');
    if (nextIndex < 0) {
      setDone(true);
    } else {
      setIndex(nextIndex);
    }
  };

  if (loading) {
    return (
      <div className="w-full min-h-[300px] flex items-center justify-center">
        <p className="text-gray-400 text-sm">Carregando curso...</p>
      </div>
    );
  }

  if (loadError || lessons.length === 0) {
    return (
      <div className="w-full py-12 text-center">
        <p className={`text-sm mb-4 ${loadError ? 'text-red-400' : 'text-gray-400'}`}>
          {loadError || 'Este curso ainda não tem lições.'}
        </p>
        <button onClick={onClose} className={primaryButton}>
          Voltar
        </button>
      </div>
    );
  }

  if (done) {
    return (
      <div className="w-full py-12 text-center">
        <h2 className="text-2xl font-bold text-emerald-400 mb-2">🎉 Curso concluído!</h2>
        <p className="text-yellow-400 font-semibold text-lg mb-1">+{course.xp} XP</p>
        <p className="text-gray-400 text-sm mb-6">
          {lessons.length} de {lessons.length} lições
        </p>
        <button onClick={onClose} className={primaryButton}>
          Voltar aos cursos
        </button>
      </div>
    );
  }

  return (
    <div className="w-full">
      <div className="flex items-center justify-between gap-4 mb-4">
        <button onClick={onClose} className="text-gray-400 hover:text-gray-200 text-sm font-medium transition">
          ← Voltar aos cursos
        </button>
        <h2 className="text-lg font-bold text-white truncate">{course.title}</h2>
        <span className="text-xs text-gray-500 whitespace-nowrap">
          {completedCount} de {lessons.length} lições
        </span>
      </div>

      <div className="h-2 bg-gray-800 rounded-full overflow-hidden mb-8">
        <div className="h-full bg-emerald-500 transition-all duration-300" style={{ width: `${percent}%` }} />
      </div>

      <p className="text-xs text-gray-500 uppercase tracking-wider mb-2">
        {lesson.unitTitle} · {lesson.taskTitle}
      </p>
      <p className="text-xl font-bold text-white mb-6">{lesson.question}</p>

      <div className="space-y-3 mb-6">
        {choices.map((choice) => {
          const isSelected = selected === choice.ID;
          let tone = 'border-gray-700 bg-gray-950 hover:border-gray-500';
          if (isSelected && result === 'wrong') tone = 'border-red-500 bg-red-950/40';
          else if (isSelected && result === 'correct') tone = 'border-emerald-500 bg-emerald-950/50';
          else if (isSelected) tone = 'border-emerald-500 bg-emerald-950/30';
          return (
            <button
              key={choice.ID}
              onClick={() => pickChoice(choice.ID)}
              disabled={locked}
              className={`w-full text-left px-4 py-3 rounded-lg border-2 text-gray-100 transition ${tone} ${
                locked && !isSelected ? 'opacity-50' : ''
              }`}
            >
              {choice.label}
            </button>
          );
        })}
      </div>

      {result === 'correct' && (
        <div className="mb-4 p-3 bg-emerald-950/50 border border-emerald-900/50 rounded-lg text-emerald-400 font-semibold text-center">
          ✅ Correto!
        </div>
      )}
      {result === 'wrong' && (
        <div className="mb-4 p-3 bg-red-950/50 border border-red-900/50 rounded-lg text-red-400 font-semibold text-center">
          ❌ Resposta incorreta, tente novamente.
        </div>
      )}
      {message && (
        <div className="mb-4 p-3 bg-red-950/50 border border-red-900/50 rounded-lg text-red-400 text-sm text-center">
          {message}
        </div>
      )}

      <div className="flex justify-end">
        {result === 'correct' ? (
          <button onClick={next} className={primaryButton}>
            Continuar
          </button>
        ) : (
          <button
            onClick={submit}
            disabled={!selected || submitting}
            className={`px-6 py-2 rounded-lg text-sm font-semibold transition ${
              selected && !submitting
                ? 'bg-emerald-600 hover:bg-emerald-500 text-white'
                : 'bg-gray-800 text-gray-500 cursor-not-allowed'
            }`}
          >
            {submitting ? 'Verificando...' : 'Verificar'}
          </button>
        )}
      </div>
    </div>
  );
}
