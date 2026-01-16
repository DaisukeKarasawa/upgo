interface StatusBadgeProps {
  state: string
  alwaysColored?: boolean
}

export default function StatusBadge({ state, alwaysColored = false }: StatusBadgeProps) {
  const stateUpper = state.toUpperCase()
  
  // Determine color class based on status
  const getColorClasses = () => {
    switch (stateUpper) {
      case 'OPEN':
        return alwaysColored 
          ? 'text-green-500' 
          : 'text-gray-400 group-hover:text-green-500'
      case 'CLOSED':
        return alwaysColored 
          ? 'text-red-500' 
          : 'text-gray-400 group-hover:text-red-500'
      case 'MERGED':
        return alwaysColored 
          ? 'text-purple-500' 
          : 'text-gray-400 group-hover:text-purple-500'
      default:
        return 'text-gray-400'
    }
  }

  const colorClasses = getColorClasses()
  const baseClass = 'text-xs font-light uppercase tracking-wider w-16 flex-shrink-0'

  return (
    <span className={`${baseClass} ${colorClasses}`}>
      {stateUpper}
    </span>
  )
}
