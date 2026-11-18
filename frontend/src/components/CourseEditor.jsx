import { useState } from 'react';
import { Input, Textarea, Select, Button } from './ui';
import {
  DEFAULT_XP,
  buildPayload,
  dropChoice,
  emptyChoice,
  emptyLesson,
  emptyModule,
  emptyTask,
  validateCourse
} from '../courseForm';

// Helpers to change one item of a list without mutating state.
const mapById = (items, id, change) => items.map((item) => (item.id === id ? change(item) : item));
const without = (items, id) => items.filter((item) => item.id !== id);

// Form where the admin builds a course: Course > Modules > Tasks > Lessons >
// Choices. `onSubmit(payload)` must resolve to true when the course was saved.
export default function CourseEditor({ onSubmit }) {
  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [category, setCategory] = useState('');
  const [difficulty, setDifficulty] = useState('beginner');
  const [xp, setXp] = useState(DEFAULT_XP);
  const [units, setUnits] = useState([]);
  const [error, setError] = useState('');
  const [submitting, setSubmitting] = useState(false);

  // Each level is nested in the previous one, so editing a lesson means
  // rebuilding module -> task -> lesson. These three do that once; everything
  // below just says what to change.
  const changeModule = (moduleId, change) => setUnits((prev) => mapById(prev, moduleId, change));
  const changeTask = (moduleId, taskId, change) =>
    changeModule(moduleId, (m) => ({ ...m, tasks: mapById(m.tasks, taskId, change) }));
  const changeLesson = (moduleId, taskId, lessonId, change) =>
    changeTask(moduleId, taskId, (t) => ({ ...t, lessons: mapById(t.lessons, lessonId, change) }));

  const addModule = () => setUnits((prev) => [...prev, emptyModule()]);
  const removeModule = (moduleId) => setUnits((prev) => without(prev, moduleId));
  const setModuleTitle = (moduleId, value) => changeModule(moduleId, (m) => ({ ...m, title: value }));

  const addTask = (moduleId) =>
    changeModule(moduleId, (m) => ({ ...m, tasks: [...m.tasks, emptyTask()] }));
  const removeTask = (moduleId, taskId) =>
    changeModule(moduleId, (m) => ({ ...m, tasks: without(m.tasks, taskId) }));
  const setTaskTitle = (moduleId, taskId, value) =>
    changeTask(moduleId, taskId, (t) => ({ ...t, title: value }));

  const addLesson = (moduleId, taskId) =>
    changeTask(moduleId, taskId, (t) => ({ ...t, lessons: [...t.lessons, emptyLesson()] }));
  const removeLesson = (moduleId, taskId, lessonId) =>
    changeTask(moduleId, taskId, (t) => ({ ...t, lessons: without(t.lessons, lessonId) }));
  const setQuestion = (moduleId, taskId, lessonId, value) =>
    changeLesson(moduleId, taskId, lessonId, (l) => ({ ...l, question: value }));

  const addChoice = (moduleId, taskId, lessonId) =>
    changeLesson(moduleId, taskId, lessonId, (l) => ({ ...l, choices: [...l.choices, emptyChoice()] }));
  const removeChoice = (moduleId, taskId, lessonId, choiceId) =>
    changeLesson(moduleId, taskId, lessonId, (l) => ({ ...l, choices: dropChoice(l.choices, choiceId) }));
  const setChoiceLabel = (moduleId, taskId, lessonId, choiceId, value) =>
    changeLesson(moduleId, taskId, lessonId, (l) => ({
      ...l,
      choices: mapById(l.choices, choiceId, (c) => ({ ...c, label: value }))
    }));
  const setCorrectChoice = (moduleId, taskId, lessonId, choiceId) =>
    changeLesson(moduleId, taskId, lessonId, (l) => ({
      ...l,
      choices: l.choices.map((c) => ({ ...c, correct: c.id === choiceId }))
    }));

  const handleSubmit = async (e) => {
    e.preventDefault();
    const form = { title, description, category, difficulty, xp, units };
    const validationError = validateCourse(form);
    setError(validationError || '');
    if (validationError) return;

    setSubmitting(true);
    const saved = await onSubmit(buildPayload(form));
    if (saved) {
      setTitle('');
      setDescription('');
      setCategory('');
      setDifficulty('beginner');
      setXp(DEFAULT_XP);
      setUnits([]);
    }
    setSubmitting(false);
  };

  return (
    <div className="w-full">
      <h1 className="text-xl font-bold text-gray-100 mb-6">Criar Novo Curso</h1>

      <section className="bg-gray-900 border border-gray-800 rounded-xl p-6 mb-6">
        <h2 className="text-sm font-bold text-gray-300 mb-4">Dados do curso</h2>
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          <Input
            label="Título"
            value={title}
            onChange={e => setTitle(e.target.value)}
            placeholder="Ex: Rust para Iniciantes"
            wrapperClassName="md:col-span-2"
          />
          <Input
            label="Categoria"
            value={category}
            onChange={e => setCategory(e.target.value)}
            placeholder="Programação"
          />
          <Select
            label="Dificuldade"
            value={difficulty}
            onChange={e => setDifficulty(e.target.value)}
          />
          <Textarea
            label="Descrição"
            value={description}
            onChange={e => setDescription(e.target.value)}
            placeholder="Resumo do curso..."
            wrapperClassName="md:col-span-3"
            rows={3}
          />
          <Input
            label="XP"
            type="number"
            value={xp}
            onChange={e => setXp(e.target.value)}
          />
        </div>
      </section>

      <section className="bg-gray-900 border border-gray-800 rounded-xl p-6 mb-6">
        <div className="flex justify-between items-center mb-4">
          <h2 className="text-sm font-bold text-gray-300">Conteúdo do curso</h2>
          <Button onClick={addModule} variant="outline">+ Módulo</Button>
        </div>

        {units.length === 0 ? (
          <p className="text-gray-500 text-sm text-center py-8">
            Adicione módulos, tarefas e lições para que os alunos possam fazer o curso. Cursos sem lições não podem ser feitos.
          </p>
        ) : (
          <div className="space-y-4">
            {units.map((module, mi) => (
              <div key={module.id} className="bg-gray-950 border border-gray-800 rounded-xl p-4">
                <div className="flex items-end justify-between gap-2 mb-3">
                  <Input
                    label="Título do módulo"
                    value={module.title}
                    onChange={e => setModuleTitle(module.id, e.target.value)}
                    wrapperClassName="flex-1"
                  />
                  <Button variant="danger" onClick={() => removeModule(module.id)}>Remover</Button>
                </div>
                <div className="ml-4 pl-4 border-l border-gray-800 space-y-3">
                  {module.tasks.map((task, ti) => (
                    <div key={task.id} className="bg-gray-950/50 border border-gray-800/50 rounded-lg p-3">
                      <div className="flex items-end justify-between gap-2 mb-2">
                        <Input
                          label="Título da tarefa"
                          value={task.title}
                          onChange={e => setTaskTitle(module.id, task.id, e.target.value)}
                          wrapperClassName="flex-1"
                        />
                        <div className="flex gap-2">
                          <Button variant="outline" onClick={() => addLesson(module.id, task.id)}>+ Lição</Button>
                          <Button variant="danger" onClick={() => removeTask(module.id, task.id)}>Remover</Button>
                        </div>
                      </div>
                      <div className="ml-4 pl-3 border-l border-gray-800/50 space-y-2">
                        {task.lessons.map((lesson, li) => (
                          <div key={lesson.id} className="bg-gray-950/30 border border-gray-800/30 rounded-lg p-3">
                            <div className="flex items-end justify-between gap-2 mb-2">
                              <Textarea
                                label="Pergunta"
                                value={lesson.question}
                                onChange={e => setQuestion(module.id, task.id, lesson.id, e.target.value)}
                                rows={2}
                                wrapperClassName="flex-1"
                              />
                              <div className="flex gap-2">
                                <Button variant="outline" onClick={() => addChoice(module.id, task.id, lesson.id)}>+ Alternativa</Button>
                                <Button variant="danger" onClick={() => removeLesson(module.id, task.id, lesson.id)}>Remover</Button>
                              </div>
                            </div>
                            <div className="space-y-2">
                              {lesson.choices.map((choice, ci) => (
                                <div key={choice.id} className="flex items-center gap-2">
                                  <input
                                    type="radio"
                                    name={`correct-${module.id}-${task.id}-${lesson.id}`}
                                    checked={choice.correct}
                                    onChange={() => setCorrectChoice(module.id, task.id, lesson.id, choice.id)}
                                    className="text-emerald-500 focus:ring-emerald-500"
                                  />
                                  <input
                                    type="text"
                                    value={choice.label}
                                    onChange={e => setChoiceLabel(module.id, task.id, lesson.id, choice.id, e.target.value)}
                                    placeholder="Alternativa"
                                    className="flex-1 bg-gray-950 border border-gray-800 rounded-lg px-3 py-1.5 text-sm focus:outline-none focus:border-emerald-500"
                                  />
                                  <Button variant="danger" onClick={() => removeChoice(module.id, task.id, lesson.id, choice.id)}>✕</Button>
                                </div>
                              ))}
                            </div>
                          </div>
                        ))}
                      </div>
                    </div>
                  ))}
                  <Button variant="outline" onClick={() => addTask(module.id)} className="w-full">+ Tarefa</Button>
                </div>
              </div>
            ))}
          </div>
        )}
      </section>

      {error && (
        <div className="bg-red-950/30 border border-red-900/50 text-red-400 rounded-lg px-4 py-3 text-sm mb-4" role="alert">
          {error}
        </div>
      )}

      <div className="flex justify-end">
        <Button variant="emerald" disabled={submitting} onClick={handleSubmit}>
          {submitting ? 'Salvando...' : 'Salvar Curso'}
        </Button>
      </div>
    </div>
  );
}