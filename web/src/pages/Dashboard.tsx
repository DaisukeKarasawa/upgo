import { useState, useEffect } from "react";
import { Link } from "react-router-dom";
import { getChanges, sync, getDashboardUpdateStatus } from "../services/api";
import StatusBadge from "../components/StatusBadge";
import ManualSyncButton from "../components/ManualSyncButton";

export default function Dashboard() {
  const [changes, setChanges] = useState<any[]>([]);
  const [loading, setLoading] = useState(false);
  const [hasUpdates, setHasUpdates] = useState(false);

  useEffect(() => {
    loadData();
    checkUpdates();

    // Poll for updates every 30 seconds
    const interval = setInterval(() => {
      checkUpdates();
    }, 30000);

    return () => clearInterval(interval);
  }, []);

  const checkUpdates = async () => {
    try {
      const status = await getDashboardUpdateStatus();
      setHasUpdates(status.has_missing_recent_changes || false);
    } catch (error) {
      console.error("Failed to check for updates", error);
      // エラー時は古い値を保持しない
      setHasUpdates(false);
    }
  };

  const loadData = async () => {
    setLoading(true);
    try {
      const changesData = await getChanges({ limit: 50 });
      setChanges(changesData.data || []);
    } catch (error) {
      console.error("Failed to fetch data", error);
    } finally {
      setLoading(false);
    }
  };

  const handleSync = async () => {
    try {
      await sync();
      // Wait a bit longer for sync to complete and DB to be updated
      setTimeout(() => {
        loadData();
        // Check for updates after sync to update badge state
        checkUpdates();
      }, 3000); // Reload after 3 seconds to ensure sync is complete
    } catch (error) {
      console.error("Failed to sync", error);
      // On sync error, also check updates to clear badge if needed
      checkUpdates();
    }
  };

  return (
    <div className="min-h-screen bg-white">
      <div className="max-w-5xl mx-auto px-6 py-12">
        <div className="flex justify-between items-start mb-12">
          <div>
            <h1 className="text-3xl font-light text-gray-900 tracking-tight mb-2">
              Upgo
            </h1>
            <p className="text-sm text-gray-400 font-light">
              Go Repository Monitoring System (Gerrit)
            </p>
          </div>
          <ManualSyncButton onSync={handleSync} hasUpdates={hasUpdates} />
        </div>

        <div>
          <div>
            {loading ? (
              <div className="text-center py-20 text-gray-400 text-sm font-light">
                Loading...
              </div>
            ) : changes.length === 0 ? (
              <div className="text-center py-20 text-gray-400 text-sm font-light">
                No changes yet. Click the sync button in the top right to fetch
                changes.
              </div>
            ) : (
              <div className="space-y-1">
                {changes.map((change) => (
                  <Link
                    key={change.id}
                    to={`/change/${change.id}`}
                    className="block py-5 px-1 hover:bg-gray-50 transition-all duration-300 ease-out group relative overflow-hidden"
                  >
                    <div className="absolute inset-0 bg-gradient-to-r from-transparent via-gray-50/50 to-transparent translate-x-[-100%] group-hover:translate-x-[100%] transition-transform duration-700 ease-in-out" />
                    <div className="flex items-start gap-3 relative">
                      <StatusBadge state={change.status} />
                      <div className="flex-1 min-w-0">
                        <h3 className="text-base font-normal text-gray-900 mb-1.5 group-hover:text-gray-700 transition-colors duration-300">
                          {change.subject}
                        </h3>
                        <p className="text-sm text-gray-500 mb-2 line-clamp-2 font-light leading-relaxed group-hover:text-gray-600 transition-colors duration-300">
                          {change.message?.substring(0, 200)}...
                        </p>
                        <div className="text-xs text-gray-400 font-light">
                          {change.owner} • {change.branch} •{" "}
                          {new Date(change.created_at).toLocaleDateString(
                            "en-US"
                          )}
                        </div>
                      </div>
                    </div>
                  </Link>
                ))}
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
