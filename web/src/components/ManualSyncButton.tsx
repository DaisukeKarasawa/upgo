import { useState } from "react";

interface ManualSyncButtonProps {
  onSync: () => Promise<void>;
  hasUpdates?: boolean;
}

export default function ManualSyncButton({ onSync, hasUpdates = false }: ManualSyncButtonProps) {
  const [syncing, setSyncing] = useState(false);

  const handleClick = async () => {
    setSyncing(true);
    try {
      await onSync();
    } finally {
      setTimeout(() => setSyncing(false), 2000);
    }
  };

  return (
    <button
      onClick={handleClick}
      disabled={syncing}
      className={`
        relative inline-flex items-center justify-center
        w-8 h-8
        transition-all duration-300 ease-out
        ${
          syncing
            ? "text-gray-300 cursor-not-allowed opacity-50"
            : "text-gray-600 hover:text-gray-900"
        }
        disabled:cursor-not-allowed
        group
      `}
      title="同期"
    >
      {hasUpdates && !syncing && (
        <span
          className="absolute top-0 left-0 w-2.5 h-2.5 bg-red-500 rounded-full z-10"
          aria-label="更新あり"
        />
      )}
      {syncing ? (
        <svg
          className="animate-spin h-4 w-4 text-gray-300"
          xmlns="http://www.w3.org/2000/svg"
          fill="none"
          viewBox="0 0 24 24"
        >
          <circle
            className="opacity-25"
            cx="12"
            cy="12"
            r="10"
            stroke="currentColor"
            strokeWidth="4"
          ></circle>
          <path
            className="opacity-75"
            fill="currentColor"
            d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
          ></path>
        </svg>
      ) : (
        <svg
          className="h-4 w-4 transition-transform duration-300 ease-out group-hover:rotate-180"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
          xmlns="http://www.w3.org/2000/svg"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={1.5}
            d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
          />
        </svg>
      )}
    </button>
  );
}
