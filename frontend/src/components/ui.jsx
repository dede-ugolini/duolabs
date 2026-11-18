// Small form building blocks shared by the course editor.

const fieldClass =
  'w-full bg-gray-950 border border-gray-800 rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-emerald-500';

export function Input({ label, wrapperClassName = '', ...props }) {
  return (
    <div className={`w-full ${wrapperClassName}`}>
      <label className="block text-xs text-gray-400 mb-1">{label}</label>
      <input className={fieldClass} {...props} />
    </div>
  );
}

export function Textarea({ label, wrapperClassName = '', rows = 3, ...props }) {
  return (
    <div className={`w-full ${wrapperClassName}`}>
      <label className="block text-xs text-gray-400 mb-1">{label}</label>
      <textarea rows={rows} className={fieldClass} {...props} />
    </div>
  );
}

export function Select({ label, wrapperClassName = '', ...props }) {
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

export function Button({ children, variant = 'emerald', className = '', ...props }) {
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
