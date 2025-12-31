import { useState, useEffect } from "react";
import { Link } from "react-router-dom";
import { getChanges, getBranches, getStatuses, sync, checkUpdates as checkUpdatesApi } from "../services/api";
import StatusBadge from "../components/StatusBadge";
import ManualSyncButton from "../components/ManualSyncButton";

interface Change {
  id: number;
  change_id: string;
  change_number: number;
  project: string;
  branch: string;
  status: string;
  subject: string;
  owner_name: string;
  owner_email?: string;
  created: string;
  updated: string;
  submitted?: string;
  last_synced_at?: string;
}

export default function Dashboard() {
  const [changes, setChanges] = useState<Change[]>([]);
  const [branches, setBranches] = useState<string[]>([]);
  const [statuses, setStatuses] = useState<string[]>([]);
  const [loading, setLoading] = useState(false);
  const [hasUpdates, setHasUpdates] = useState(false);
  const [selectedBranch, setSelectedBranch] = useState<string>("");
  const [selectedStatus, setSelectedStatus] = useState<string>("");

  useEffect(() => {
    loadFilters();
    loadData();
    checkUpdates();

    const interval = setInterval(() => {
      checkUpdates();
    }, 30000);

    return () => clearInterval(interval);
  }, []);

  useEffect(() => {
    loadData();
  }, [selectedBranch, selectedStatus]);

  const loadFilters = async () => {
    try {
      const [branchesData, statusesData] = await Promise.all([
        getBranches(),
        getStatuses(),
      ]);
      setBranches(branchesData.branches || []);
      setStatuses(statusesData.statuses || []);
    } catch (error) {
      console.error("Failed to load filters", error);
    }
  };

  const checkUpdates = async () => {
    try {
      const status = await checkUpdatesApi();
      setHasUpdates(status.has_updates || false);
    } catch (error) {
      console.error("Failed to check for updates", error);
    }
  };

  const loadData = async () => {
    setLoading(true);
    try {
      const params: { limit: number; branch?: string; status?: string } = { limit: 50 };
      if (selectedBranch) params.branch = selectedBranch;
      if (selectedStatus) params.status = selectedStatus;
      const changesData = await getChanges(params);
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
      setTimeout(() => {
        loadData();
        checkUpdates();
      }, 2000);
    } catch (error) {
      console.error("Failed to sync", error);
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
              Go Gerrit Change Monitoring System
            </p>
          </div>
          <ManualSyncButton onSync={handleSync} hasUpdates={hasUpdates} />
        </div>

        <div className="flex gap-4 mb-8">
          <select
            value={selectedBranch}
            onChange={(e) => setSelectedBranch(e.target.value)}
            className="px-3 py-2 border border-gray-200 rounded text-sm font-light text-gray-700 focus:outline-none focus:border-gray-400"
          >
            <option value="">All Branches</option>
            {branches.map((branch) => (
              <option key={branch} value={branch}>{branch}</option>
            ))}
          </select>
          <select
            value={selectedStatus}
            onChange={(e) => setSelectedStatus(e.target.value)}
            className="px-3 py-2 border border-gray-200 rounded text-sm font-light text-gray-700 focus:outline-none focus:border-gray-400"
          >
            <option value="">All Statuses</option>
            {statuses.map((status) => (
              <option key={status} value={status}>{status}</option>
            ))}
          </select>
        </div>

        <div>
          <div>
            {loading ? (
              <div className="text-center py-20 text-gray-400 text-sm font-light">
                Loading...
              </div>
            ) : changes.length === 0 ? (
              <div className="text-center py-20 text-gray-400 text-sm font-light">
                No changes found. Click "Sync" to fetch changes from Gerrit.
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
                          <span className="text-gray-400 mr-2">#{change.change_number}</span>
                          {change.subject}
                        </h3>
                        <div className="text-xs text-gray-400 font-light flex items-center gap-2">
                          <span>{change.owner_name}</span>
                          <span>•</span>
                          <span>{change.branch}</span>
                          <span>•</span>
                          <span>{new Date(change.updated).toLocaleDateString("en-US")}</span>
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
