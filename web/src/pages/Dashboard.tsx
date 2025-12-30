import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { getPRs, getIssues, sync } from '../services/api'
import StatusBadge from '../components/StatusBadge'
import ManualSyncButton from '../components/ManualSyncButton'

export default function Dashboard() {
  const [prs, setPRs] = useState<any[]>([])
  const [issues, setIssues] = useState<any[]>([])
  const [activeTab, setActiveTab] = useState<'prs' | 'issues'>('prs')
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    loadData()
  }, [])

  const loadData = async () => {
    setLoading(true)
    try {
      const [prsData, issuesData] = await Promise.all([
        getPRs({ limit: 50 }),
        getIssues({ limit: 50 }),
      ])
      setPRs(prsData.data || [])
      setIssues(issuesData.data || [])
    } catch (error) {
      console.error('データの取得に失敗しました', error)
    } finally {
      setLoading(false)
    }
  }

  const handleSync = async () => {
    try {
      await sync()
      setTimeout(loadData, 2000) // 2秒後に再読み込み
    } catch (error) {
      console.error('同期に失敗しました', error)
    }
  }

  return (
    <div className="min-h-screen bg-white">
      <div className="max-w-5xl mx-auto px-6 py-12">
        <div className="flex justify-between items-start mb-12">
          <div>
            <h1 className="text-3xl font-light text-gray-900 tracking-tight mb-2">
              UpGo
            </h1>
            <p className="text-sm text-gray-400 font-light">
              Goリポジトリ監視システム
            </p>
          </div>
          <ManualSyncButton onSync={handleSync} />
        </div>

        <div>
          <nav className="flex gap-8 mb-8 border-b border-gray-100">
            <button
              onClick={() => setActiveTab('prs')}
              className={`pb-3 text-sm font-light transition-all duration-300 ease-out relative ${
                activeTab === 'prs'
                  ? 'text-gray-900'
                  : 'text-gray-400 hover:text-gray-600'
              }`}
            >
              Pull Requests
              {activeTab === 'prs' && (
                <span className="absolute bottom-0 left-0 right-0 h-px bg-gray-900 animate-[slideIn_0.3s_ease-out]" />
              )}
            </button>
            <button
              onClick={() => setActiveTab('issues')}
              className={`pb-3 text-sm font-light transition-all duration-300 ease-out relative ${
                activeTab === 'issues'
                  ? 'text-gray-900'
                  : 'text-gray-400 hover:text-gray-600'
              }`}
            >
              Issues
              {activeTab === 'issues' && (
                <span className="absolute bottom-0 left-0 right-0 h-px bg-gray-900 animate-[slideIn_0.3s_ease-out]" />
              )}
            </button>
          </nav>

          <div>
            {loading ? (
              <div className="text-center py-20 text-gray-400 text-sm font-light">読み込み中...</div>
            ) : activeTab === 'prs' ? (
              <div className="space-y-1">
                {prs.map((pr) => (
                  <Link
                    key={pr.id}
                    to={`/pr/${pr.id}`}
                    className="block py-5 px-1 hover:bg-gray-50 transition-all duration-300 ease-out group relative overflow-hidden"
                  >
                    <div className="absolute inset-0 bg-gradient-to-r from-transparent via-gray-50/50 to-transparent translate-x-[-100%] group-hover:translate-x-[100%] transition-transform duration-700 ease-in-out" />
                    <div className="flex items-start gap-3 relative">
                      <StatusBadge state={pr.state} />
                      <div className="flex-1 min-w-0">
                        <h3 className="text-base font-normal text-gray-900 mb-1.5 group-hover:text-gray-700 transition-colors duration-300">
                          {pr.title}
                        </h3>
                        <p className="text-sm text-gray-500 mb-2 line-clamp-2 font-light leading-relaxed group-hover:text-gray-600 transition-colors duration-300">
                          {pr.body?.substring(0, 200)}...
                        </p>
                        <div className="text-xs text-gray-400 font-light">
                          {pr.author} • {new Date(pr.created_at).toLocaleDateString('ja-JP')}
                        </div>
                      </div>
                    </div>
                  </Link>
                ))}
              </div>
            ) : (
              <div className="space-y-1">
                {issues.map((issue) => (
                  <Link
                    key={issue.id}
                    to={`/issue/${issue.id}`}
                    className="block py-5 px-1 hover:bg-gray-50 transition-all duration-300 ease-out group relative overflow-hidden"
                  >
                    <div className="absolute inset-0 bg-gradient-to-r from-transparent via-gray-50/50 to-transparent translate-x-[-100%] group-hover:translate-x-[100%] transition-transform duration-700 ease-in-out" />
                    <div className="flex items-start gap-3 relative">
                      <StatusBadge state={issue.state} />
                      <div className="flex-1 min-w-0">
                        <h3 className="text-base font-normal text-gray-900 mb-1.5 group-hover:text-gray-700 transition-colors duration-300">
                          {issue.title}
                        </h3>
                        <p className="text-sm text-gray-500 mb-2 line-clamp-2 font-light leading-relaxed group-hover:text-gray-600 transition-colors duration-300">
                          {issue.body?.substring(0, 200)}...
                        </p>
                        <div className="text-xs text-gray-400 font-light">
                          {issue.author} • {new Date(issue.created_at).toLocaleDateString('ja-JP')}
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
  )
}
