import { useEffect, useMemo, useState } from 'react'
import { Activity, Copy, ExternalLink, Link2, Plus, RefreshCw, Search, Send } from 'lucide-react'
import { getEvents, getLinks } from '../api.js'
import EventsTable from '../components/EventsTable.jsx'
import LinkModal from '../components/LinkModal.jsx'

export default function Dashboard() {
  const [links, setLinks] = useState([])
  const [events, setEvents] = useState([])
  const [filters, setFilters] = useState({ pixel_id: '', from: '', to: '' })
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [modalOpen, setModalOpen] = useState(false)
  const [copied, setCopied] = useState('')

  async function load() {
    setLoading(true)
    setError('')
    try {
      const [nextLinks, nextEvents] = await Promise.all([getLinks(), getEvents(filters)])
      setLinks(nextLinks)
      setEvents(nextEvents)
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    load()
  }, [])

  useEffect(() => {
    const timer = setTimeout(() => {
      getEvents(filters).then(setEvents).catch((err) => setError(err.message))
    }, 250)
    return () => clearTimeout(timer)
  }, [filters])

  const stats = useMemo(() => {
    const uniquePixels = new Set(links.map((link) => link.pixel_id)).size
    const identified = events.filter((event) => event.uid || event.email).length
    return { uniquePixels, identified }
  }, [links, events])

  async function copy(value) {
    await navigator.clipboard.writeText(value)
    setCopied(value)
    setTimeout(() => setCopied(''), 1400)
  }

  function updateFilter(key, value) {
    setFilters((current) => ({ ...current, [key]: value }))
  }

  return (
    <main className="shell">
      <header className="topbar">
        <div className="brand-lockup">
          <div className="brand-mark">PM</div>
          <div>
            <h1>PixelMapper</h1>
            <p>Meta Pixel link mapper for known users</p>
          </div>
        </div>
        <div className="topbar-actions">
          <button className="icon-button" type="button" onClick={load} title="Refresh" aria-label="Refresh">
            <RefreshCw size={18} />
          </button>
          <button className="primary-button" type="button" onClick={() => setModalOpen(true)}>
            <Plus size={18} />
            Create Link
          </button>
        </div>
      </header>

      <section className="stats-grid" aria-label="PixelMapper summary">
        <Metric icon={<Link2 size={20} />} label="Tracking links" value={links.length} />
        <Metric icon={<Activity size={20} />} label="Tracked events" value={events.length} />
        <Metric icon={<Send size={20} />} label="Pixels mapped" value={stats.uniquePixels} />
        <Metric icon={<Search size={20} />} label="Identified users" value={stats.identified} />
      </section>

      {error ? <div className="notice error">{error}</div> : null}

      <section className="workspace">
        <aside className="links-panel">
          <div className="panel-heading">
            <h2>Links</h2>
            <button className="ghost-button" type="button" onClick={() => setModalOpen(true)}>
              <Plus size={16} />
              New
            </button>
          </div>
          <div className="link-list">
            {links.length === 0 && !loading ? (
              <div className="empty-state">Create first tracking link.</div>
            ) : null}
            {links.map((link) => (
              <article className="link-card" key={link.token}>
                <div className="link-card-top">
                  <strong>{link.label || 'Untitled campaign'}</strong>
                  <span>{link.event_count || 0} hits</span>
                </div>
                <div className="mono">{link.pixel_id}</div>
                <div className="link-actions">
                  <button className="icon-button" type="button" onClick={() => copy(link.tracking_url)} title="Copy link" aria-label="Copy link">
                    <Copy size={16} />
                  </button>
                  <a className="icon-button" href={link.tracking_url} target="_blank" rel="noreferrer" title="Open link" aria-label="Open link">
                    <ExternalLink size={16} />
                  </a>
                  <span className="copy-note">{copied === link.tracking_url ? 'Copied' : link.token}</span>
                </div>
              </article>
            ))}
          </div>
        </aside>

        <section className="events-panel">
          <div className="panel-heading events-heading">
            <h2>Tracked Events</h2>
            <div className="filters">
              <input
                type="search"
                placeholder="Pixel ID"
                value={filters.pixel_id}
                onChange={(event) => updateFilter('pixel_id', event.target.value)}
              />
              <input type="date" value={filters.from} onChange={(event) => updateFilter('from', event.target.value)} />
              <input type="date" value={filters.to} onChange={(event) => updateFilter('to', event.target.value)} />
            </div>
          </div>
          <EventsTable events={events} loading={loading} onCopy={copy} copied={copied} />
        </section>
      </section>

      <LinkModal
        open={modalOpen}
        onClose={() => setModalOpen(false)}
        onCreated={() => {
          setModalOpen(false)
          load()
        }}
      />
    </main>
  )
}

function Metric({ icon, label, value }) {
  return (
    <article className="metric">
      <div className="metric-icon">{icon}</div>
      <div>
        <span>{label}</span>
        <strong>{value}</strong>
      </div>
    </article>
  )
}
