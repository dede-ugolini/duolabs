// Pure helpers behind CourseEditor: blank items, validation and the payload
// sent to the API. No React in here, so each piece is easy to read on its own.

export const DEFAULT_XP = 500;
export const MAX_XP = 32767; // the API stores xp as int16

let lastId = 0;
// Local ids only identify an item while it is being edited (React keys); the
// API assigns the real ones.
const newId = () => ++lastId;

export const emptyChoice = (correct = false) => ({ id: newId(), label: '', correct });

export const emptyLesson = () => ({
  id: newId(),
  question: '',
  choices: [emptyChoice(true), emptyChoice()]
});

export const emptyTask = () => ({ id: newId(), title: '', lessons: [emptyLesson()] });

export const emptyModule = () => ({ id: newId(), title: '', tasks: [emptyTask()] });

// Removes a choice, keeping exactly one correct choice (the first one when the
// removed choice was the correct one).
export function dropChoice(choices, choiceId) {
  const rest = choices.filter((c) => c.id !== choiceId);
  return rest.some((c) => c.correct) ? rest : rest.map((c, i) => ({ ...c, correct: i === 0 }));
}

// A blank XP field means "use the default"; otherwise it must be a whole
// number the API can store. Returns null when invalid.
export function parseXp(xp) {
  const value = xp === '' ? DEFAULT_XP : Number(xp);
  return Number.isInteger(value) && value >= 0 && value <= MAX_XP ? value : null;
}

function validateLesson(lesson, where) {
  if (!lesson.question.trim()) return `${where}: informe a pergunta.`;
  if (lesson.choices.length < 2) return `${where}: mínimo 2 alternativas.`;
  const blank = lesson.choices.findIndex((c) => !c.label.trim());
  if (blank >= 0) return `${where} · Alternativa ${blank + 1}: informe o texto da alternativa.`;
  if (lesson.choices.filter((c) => c.correct).length !== 1) {
    return `${where}: marque exatamente uma alternativa como correta.`;
  }
  return null;
}

// Returns the first problem found as a message for the user, or null when the
// course can be sent. Mirrors the checks the API does.
export function validateCourse({ title, description, category, xp, units }) {
  if (!title.trim()) return 'Informe o título do curso.';
  if (!description.trim()) return 'Informe a descrição do curso.';
  if (!category.trim()) return 'Informe a categoria do curso.';
  if (parseXp(xp) === null) return `Informe um XP inteiro entre 0 e ${MAX_XP}.`;

  for (const [mi, module] of units.entries()) {
    const moduleWhere = `Módulo ${mi + 1}`;
    if (!module.title.trim()) return `${moduleWhere}: informe o título do módulo.`;

    for (const [ti, task] of module.tasks.entries()) {
      const taskWhere = `${moduleWhere} · Tarefa ${ti + 1}`;
      if (!task.title.trim()) return `${taskWhere}: informe o título da tarefa.`;

      for (const [li, lesson] of task.lessons.entries()) {
        const error = validateLesson(lesson, `${taskWhere} · Lição ${li + 1}`);
        if (error) return error;
      }
    }
  }
  return null;
}

// Turns the form state into the JSON the API expects (no local ids, trimmed).
export function buildPayload({ title, description, category, difficulty, xp, units }) {
  return {
    title: title.trim(),
    description: description.trim(),
    category: category.trim(),
    difficulty,
    xp: parseXp(xp),
    units: units.map((m) => ({
      title: m.title.trim(),
      tasks: m.tasks.map((t) => ({
        title: t.title.trim(),
        lessons: t.lessons.map((l) => ({
          question: l.question.trim(),
          choices: l.choices.map((c) => ({ label: c.label.trim(), correct: c.correct }))
        }))
      }))
    }))
  };
}
