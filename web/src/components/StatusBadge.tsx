interface StatusBadgeProps {
  state: string
}

export default function StatusBadge({ state }: StatusBadgeProps) {
  const getColor = () => {
    switch (state) {
      case 'open':
        return 'bg-green-100 text-green-800'
      case 'merged':
        return 'bg-blue-100 text-blue-800'
      case 'closed':
        return 'bg-gray-100 text-gray-800'
      default:
        return 'bg-gray-100 text-gray-800'
    }
  }

  return (
    <span className={`px-2 py-1 text-xs font-semibold rounded ${getColor()}`}>
      {state}
    </span>
  )
}
