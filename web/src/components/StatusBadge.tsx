interface StatusBadgeProps {
  state: string
}

export default function StatusBadge({ state }: StatusBadgeProps) {
  return (
    <span className="text-xs text-gray-400 font-light uppercase tracking-wider">
      {state}
    </span>
  )
}
