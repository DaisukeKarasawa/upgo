interface StatusBadgeProps {
  state: string
  alwaysColored?: boolean
}

export default function StatusBadge({ state, alwaysColored = false }: StatusBadgeProps) {
  const stateUpper = state.toUpperCase()
  
  // Determine color class based on status
  // Supports both GitHub PR states (open/closed/merged) and Gerrit Change states (NEW/MERGED/ABANDONED)
  const getColorClasses = () => {
    switch (stateUpper) {
      case 'OPEN':
      case 'NEW':
        return alwaysColored 
          ? 'text-green-500' 
          : 'text-gray-400 group-hover:text-green-500'
      case 'CLOSED':
      case 'ABANDONED':
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

  // Format status text to match Gerrit UI display
  // Gerrit UI shows "Active" for NEW status, "Merged" for MERGED, "Abandoned" for ABANDONED
  const formatStatusText = (status: string): string => {
    const upper = status.toUpperCase()
    switch (upper) {
      case 'NEW':
        return 'Active' // Gerrit UI displays "Active" for NEW status
      case 'MERGED':
        return 'Merged'
      case 'ABANDONED':
        return 'Abandoned'
      case 'OPEN':
        return 'Open'
      case 'CLOSED':
        return 'Closed'
      default:
        // Fallback: capitalize first letter, lowercase rest
        return status.charAt(0).toUpperCase() + status.slice(1).toLowerCase()
    }
  }

  const colorClasses = getColorClasses()
  const baseClass = 'text-xs font-light tracking-wider w-16 flex-shrink-0'
  const displayText = formatStatusText(state)

  return (
    <span className={`${baseClass} ${colorClasses}`}>
      {displayText}
    </span>
  )
}
