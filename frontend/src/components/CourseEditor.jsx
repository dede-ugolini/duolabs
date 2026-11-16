import React, { useState, useRef } from 'react';

const fieldClass =
  'w-full bg-gray-950 border border-gray-800 rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-emerald-500';

function Input({ label, wrapperClassName = '', ...props }) {
  return (
    <div className={`w-full ${wrapperClassName}`}>
      <label className="block text-xs text-gray-400 mb-1">{label}</label>
      <input className={fieldClass} {...props} />
    </div>
  );
}

function Textarea({ label, wrapperClassName = '', rows = 3, ...props }) {
  return (
    <div className={`w-full ${wrapperClassName}`}>
      <label className="block text-xs text-gray-400 mb-1">{label}</label>
      <textarea rows={rows} className={fieldClass} {...props} />
    </div>
  );
}

function Select({ label, wrapperClassName = '', ...props }) {
  return (
    <div className={`w-full ${wrapperClassName}`}>
      <label className="block text-xs text-gray-400 mb-1">{label}</label>
      <select className={fieldClass} {...props}>
        <option value="beginner">Beginner</option>
        <option value="intermediate">Intermediate</option>
        <option value="advanced">Advanced</option>
      </select>
    </div>
  );
}

const buttonVariants = {
  emerald: 'bg-emerald-600 hover:bg-emerald-500 text-white',
  outline: 'bg-gray-950 border border-gray-700 text-gray-300 hover:border-gray-500',
  danger: 'bg-red-950/40 hover:bg-red-900/50 text-red-400 border border-red-900/60'
};

function Button({ children, variant = 'emerald', className = '', ...props }) {
  return (
    <button
      type="button"
      className={`px-4 py-2 rounded-lg text-sm font-semibold transition disabled:opacity-50 disabled:cursor-not-allowed ${buttonVariants[variant]} ${className}`}
      {...props}
    >
      {children}
    </button>
  );
}

function dropChoice(choices, choiceId) {
  const rest = choices.filter(c => c.id !== choiceId);
  return rest.some(c => c.correct) ? rest : rest.map((c, i) => ({ ...c, correct: i === 0 }));
}

export default function CourseEditor({ onSubmit }) {
  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [category, setCategory] = useState('');
  const [difficulty, setDifficulty] = useState('beginner');
  const [xp, setXp] = useState(500);
  const [units, setUnits] = useState([]);
  const [error, setError] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const idCounter = useRef(0);

  const newId = () => {
    idCounter.current += 1;
    return idCounter.current;
  };

  const emptyModule = () => ({
    id: newId(),
    title: '',
    tasks: [emptyTask()]
  });

  const emptyTask = () => ({
    id: newId(),
    title: '',
    lessons: [emptyLesson()]
  });

  const emptyLesson = () => ({
    id: newId(),
    question: '',
    choices: [
      { id: newId(), label: '', correct: true },
      { id: newId(), label: '', correct: false }
    ]
  });

  const addModule = () => setUnits(prev => [...prev, emptyModule()]);

  const removeModule = (moduleId) => setUnits(prev => prev.filter(m => m.id !== moduleId));

  const updateModuleTitle = (moduleId, value) => setUnits(prev => prev.map(m => m.id === moduleId ? { ...m, title: value } : m));

  const addTask = (moduleId) => setUnits(prev => prev.map(m => m.id === moduleId ? { ...m, tasks: [...m.tasks, emptyTask()] } : m));

  const removeTask = (moduleId, taskId) => setUnits(prev => prev.map(m => m.id === moduleId ? { ...m, tasks: m.tasks.filter(t => t.id !== taskId) } : m));

  const updateTaskTitle = (moduleId, taskId, value) => setUnits(prev => prev.map(m => m.id === moduleId ? { ...m, tasks: m.tasks.map(t => t.id === taskId ? { ...t, title: value } : t) } : m));

  const addLesson = (moduleId, taskId) => setUnits(prev => prev.map(m => m.id === moduleId ? { ...m, tasks: m.tasks.map(t => t.id === taskId ? { ...t, lessons: [...t.lessons, emptyLesson()] } : t) } : m));

  const removeLesson = (moduleId, taskId, lessonId) => setUnits(prev => prev.map(m => m.id === moduleId ? { ...m, tasks: m.tasks.map(t => t.id === taskId ? { ...t, lessons: t.lessons.filter(l => l.id !== lessonId) } : t) } : m));

  const updateLessonQuestion = (moduleId, taskId, lessonId, value) => setUnits(prev => prev.map(m => m.id === moduleId ? { ...m, tasks: m.tasks.map(t => t.id === taskId ? { ...t, lessons: t.lessons.map(l => l.id === lessonId ? { ...l, question: value } : l) } : t) } : m));

  const addChoice = (moduleId, taskId, lessonId) => setUnits(prev => prev.map(m => m.id === moduleId ? { ...m, tasks: m.tasks.map(t => t.id === taskId ? { ...t, lessons: t.lessons.map(l => l.id === lessonId ? { ...l, choices: [...l.choices, { id: newId(), label: '', correct: false }] } : l) } : t) } : m));

  const removeChoice = (moduleId, taskId, lessonId, choiceId) => setUnits(prev => prev.map(m => m.id === moduleId ? { ...m, tasks: m.tasks.map(t => t.id === taskId ? { ...t, lessons: t.lessons.map(l => l.id === lessonId ? { ...l, choices: dropChoice(l.choices, choiceId) } : l) } : t) } : m));

  const updateChoice = (moduleId, taskId, lessonId, choiceId, field, value) => setUnits(prev => prev.map(m => m.id === moduleId ? { ...m, tasks: m.tasks.map(t => t.id === taskId ? { ...t, lessons: t.lessons.map(l => l.id === lessonId ? { ...l, choices: l.choices.map(c => c.id === choiceId ? { ...c, [field]: value } : c) } : l) } : t) } : m));

  const setCorrectChoice = (moduleId, taskId, lessonId, choiceId) => setUnits(prev => prev.map(m => m.id === moduleId ? { ...m, tasks: m.tasks.map(t => t.id === taskId ? { ...t, lessons: t.lessons.map(l => l.id === lessonId ? { ...l, choices: l.choices.map(c => ({ ...c, correct: c.id === choiceId })) } : l) } : t) } : m));

  const validate = () => {
    const t = title.trim();
    const d = description.trim();
    const c = category.trim();
    if (!t) return 'Informe o título do curso.';
    if (!d) return 'Informe a descrição do curso.';
    if (!c) return 'Informe a categoria do curso.';

    for (let mi = 0; mi < units.length; mi++) {
      const m = units[mi];
      const mTitle = m.title.trim();
      if (!mTitle) return `Módulo ${mi + 1}: informe o título do módulo.`;

      for (let ti = 0; ti < m.tasks.length; ti++) {
        const task = m.tasks[ti];
        const taskTitle = task.title.trim();
        if (!taskTitle) return `Módulo ${mi + 1} · Tarefa ${ti + 1}: informe o título da tarefa.`;

        for (let li = 0; li < task.lessons.length; li++) {
          const lesson = task.lessons[li];
          const q = lesson.question.trim();
          if (!q) return `Módulo ${mi + 1} · Tarefa ${ti + 1} · Lição ${li + 1}: informe a pergunta.`;
          if (lesson.choices.length < 2) return `Módulo ${mi + 1} · Tarefa ${ti + 1} · Lição ${li + 1}: mínimo 2 alternativas.`;
          let correctCount = 0;
          for (let ci = 0; ci < lesson.choices.length; ci++) {
            const choice = lesson.choices[ci];
            if (!choice.label.trim()) return `Módulo ${mi + 1} · Tarefa ${ti + 1} · Lição ${li + 1} · Alternativa ${ci + 1}: informe o texto da alternativa.`;
            if (choice.correct) correctCount++;
          }
          if (correctCount !== 1) return `Módulo ${mi + 1} · Tarefa ${ti + 1} · Lição ${li + 1}: marque exatamente uma alternativa como correta.`;
        }
      }
    }
    return null;
  };

  const buildPayload = () => ({
    title: title.trim(),
    description: description.trim(),
    category: category.trim(),
    difficulty,
    xp: parseInt(xp) || 500,
    units: units.map(m => ({
      title: m.title.trim(),
      tasks: m.tasks.map(t => ({
        title: t.title.trim(),
        lessons: t.lessons.map(l => ({
          question: l.question.trim(),
          choices: l.choices.map(c => ({ label: c.label.trim(), correct: c.correct }))
        }))
      }))
    }))
  });

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');
    const validationError = validate();
    if (validationError) {
      setError(validationError);
      return;
    }
    setSubmitting(true);
    const payload = buildPayload();
    const ok = await onSubmit(payload);
    if (ok) {
      setTitle('');
      setDescription('');
      setCategory('');
      setDifficulty('beginner');
      setXp(500);
      setUnits([]);
      idCounter.current = 0;
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
                    onChange={e => updateModuleTitle(module.id, e.target.value)}
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
                          onChange={e => updateTaskTitle(module.id, task.id, e.target.value)}
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
                                onChange={e => updateLessonQuestion(module.id, task.id, lesson.id, e.target.value)}
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
                                    onChange={e => updateChoice(module.id, task.id, lesson.id, choice.id, 'label', e.target.value)}
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