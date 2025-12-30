import { useState } from "react";

interface ManualSyncButtonProps {
  onSync: () => Promise<void>;
}

export default function ManualSyncButton({ onSync }: ManualSyncButtonProps) {
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
      className={`px-4 py-2 rounded-md font-medium ${
        syncing
          ? "bg-gray-400 cursor-not-allowed"
          : "bg-blue-600 hover:bg-blue-700 text-white"
      }`}
    >
      {syncing ? "同期中..." : "同期"}
    </button>
  );
}
