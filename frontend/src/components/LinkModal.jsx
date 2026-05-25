import { useState } from 'react'
import { Copy, Link2, X } from 'lucide-react'
import { createLink } from '../api.js'

export default function LinkModal({ open, onClose, onCreated }) {
  const [form, setForm] = useState({ pixel_id: '', label: '', redirect_url: '' })
  const [created, setCreated] = useState(null)
  const [error, setError] = useState('')
  const [saving, setSaving] = useState(false)

  if (!open) return null

  function update(key, value) {
    setForm((current) => ({ ...current, [key]: value }))
  }

  async function submit(event) {
    event.preventDefault()
    setSaving(true)
    setError('')
    setCreated(null)
    try {
      const result = await createLink(form)
      setCreated(result)
    } catch (err) {
      setError(err.message)
    } finally {
      setSaving(false)
    }
  }

  async function copyLink() {
    if (created?.tracking_url) {
      await navigator.clipboard.writeText(created.tracking_url)
    }
  }

  function close() {
    setForm({ pixel_id: '', label: '', redirect_url: '' })
    setCreated(null)
    setError('')
    onCreated()
  }

  return (
    <div className="modal-backdrop" role="presentation" onMouseDown={onClose}>
      <section className="modal" role="dialog" aria-modal="true" aria-labelledby="create-link-title" onMouseDown={(event) => event.stopPropagation()}>
        <div className="modal-heading">
          <div>
            <div className="modal-icon"><Link2 size={20} /></div>
            <h2 id="create-link-title">Create Tracking Link</h2>
          </div>
          <button className="icon-button" type="button" onClick={onClose} title="Close" aria-label="Close">
            <X size={18} />
          </button>
        </div>

        <form onSubmit={submit} className="form">
          <label>
            <span>Meta Pixel ID</span>
            <input
              required
              inputMode="numeric"
              pattern="[0-9]+"
              placeholder="123456789012345"
              value={form.pixel_id}
              onChange={(event) => update('pixel_id', event.target.value)}
            />
          </label>
          <label>
            <span>Campaign label</span>
            <input placeholder="Summer Campaign" value={form.label} onChange={(event) => update('label', event.target.value)} />
          </label>
          <label>
            <span>Redirect URL</span>
            <input type="url" placeholder="https://mybrand.com" value={form.redirect_url} onChange={(event) => update('redirect_url', event.target.value)} />
          </label>

          {error ? <div className="notice error">{error}</div> : null}

          {created ? (
            <div className="created-box">
              <span>Tracking URL</span>
              <strong>{created.tracking_url}</strong>
              <button className="ghost-button" type="button" onClick={copyLink}>
                <Copy size={16} />
                Copy
              </button>
            </div>
          ) : null}

          <div className="modal-actions">
            <button className="ghost-button" type="button" onClick={created ? close : onClose}>
              {created ? 'Done' : 'Cancel'}
            </button>
            <button className="primary-button" type="submit" disabled={saving}>
              {saving ? 'Creating...' : 'Create Link'}
            </button>
          </div>
        </form>
      </section>
    </div>
  )
}
