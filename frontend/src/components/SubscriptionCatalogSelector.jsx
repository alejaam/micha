import { useCallback, useEffect, useMemo, useState } from 'react'
import { BottomSheet } from './BottomSheet'
import { listSubscriptionServices, linkCatalogService, unlinkCatalogService } from '../api'
import { useAuth } from '../context/AuthContext'
import { formatCurrency } from '../utils'

/**
 * SubscriptionCatalogSelector — modal on desktop, BottomSheet on mobile.
 *
 * @param {Object} props
 * @param {boolean} props.open
 * @param {string} props.recurringExpenseId
 * @param {string[]} props.linkedServiceIds  — currently linked service IDs
 * @param {Function} props.onClose
 * @param {Function} props.onSave  — called after changes are persisted
 */
export function SubscriptionCatalogSelector({
    open,
    recurringExpenseId,
    linkedServiceIds = [],
    onClose,
    onSave,
}) {
    const { handleProtectedError } = useAuth()
    const [services, setServices] = useState([])
    const [loading, setLoading] = useState(false)
    const [saving, setSaving] = useState(false)
    const [localLinks, setLocalLinks] = useState(new Set(linkedServiceIds))
    const [customPrices, setCustomPrices] = useState({})
    const [error, setError] = useState(null)

    // Reset local state when modal opens with new data
    useEffect(() => {
        if (open) {
            setLocalLinks(new Set(linkedServiceIds))
            setCustomPrices({})
            setError(null)
        }
    }, [open, linkedServiceIds])

    // Fetch catalog services
    useEffect(() => {
        if (!open) return
        let cancelled = false
        setLoading(true)
        listSubscriptionServices()
            .then((data) => {
                if (!cancelled) setServices(Array.isArray(data) ? data : [])
            })
            .catch((err) => {
                if (!cancelled) {
                    handleProtectedError(err)
                    setError('Error al cargar servicios')
                }
            })
            .finally(() => {
                if (!cancelled) setLoading(false)
            })
        return () => { cancelled = true }
    }, [open, handleProtectedError])

    const handleToggle = useCallback(async (serviceId) => {
        setSaving(true)
        setError(null)
        try {
            const wasLinked = localLinks.has(serviceId)
            if (wasLinked) {
                await unlinkCatalogService({
                    recurringExpenseId,
                    catalogServiceId: serviceId,
                })
                setLocalLinks((prev) => {
                    const next = new Set(prev)
                    next.delete(serviceId)
                    return next
                })
                // Clear custom price for this service
                setCustomPrices((prev) => {
                    const next = { ...prev }
                    delete next[serviceId]
                    return next
                })
            } else {
                const customPrice = customPrices[serviceId]
                await linkCatalogService({
                    recurringExpenseId,
                    catalogServiceId: serviceId,
                    customPriceCents: typeof customPrice === 'number' && customPrice > 0 ? customPrice : undefined,
                })
                setLocalLinks((prev) => new Set(prev).add(serviceId))
            }
            onSave?.()
        } catch (err) {
            handleProtectedError(err)
            setError(err.message ?? 'Error al cambiar estado del servicio')
        } finally {
            setSaving(false)
        }
    }, [recurringExpenseId, localLinks, customPrices, onSave, handleProtectedError])

    const handleCustomPriceChange = useCallback((serviceId, value) => {
        const num = Number(value)
        setCustomPrices((prev) => ({
            ...prev,
            [serviceId]: Number.isNaN(num) ? 0 : num,
        }))
    }, [])

    const bundles = useMemo(() => services.filter((s) => s.is_bundle), [services])
    const standaloneServices = useMemo(() => services.filter((s) => !s.is_bundle), [services])

    const renderServiceList = (items) => (
        <div className="catalogServiceList">
            {items.map((service) => {
                const isLinked = localLinks.has(service.id)
                const customPrice = customPrices[service.id]
                return (
                    <label key={service.id} className={`catalogServiceItem ${isLinked ? 'catalogServiceItem--linked' : ''}`}>
                        <div className="catalogServiceItemMain">
                            <input
                                type="checkbox"
                                checked={isLinked}
                                onChange={() => handleToggle(service.id)}
                                disabled={saving}
                                className="catalogServiceCheckbox"
                            />
                            <div className="catalogServiceInfo">
                                <span className="catalogServiceName">{service.name}</span>
                                {service.is_bundle && (
                                    <span className="catalogServiceBadge">Bundle</span>
                                )}
                                <span className="catalogServicePrice">
                                    {formatCurrency(service.standalone_price_cents, service.currency)}
                                </span>
                            </div>
                        </div>
                        {isLinked && (
                            <div className="catalogServiceOverride">
                                <label className="catalogServiceOverrideLabel">
                                    Precio personalizado:
                                    <input
                                        type="number"
                                        className="input catalogServiceOverrideInput"
                                        min="0"
                                        step="1"
                                        placeholder={String(service.standalone_price_cents)}
                                        value={customPrice > 0 ? customPrice : ''}
                                        onChange={(e) => handleCustomPriceChange(service.id, e.target.value)}
                                        disabled={saving}
                                        onClick={(e) => e.stopPropagation()}
                                    />
                                </label>
                            </div>
                        )}
                    </label>
                )
            })}
        </div>
    )

    const content = (
        <>
            {error && <p className="u-text-sm u-text-danger">{error}</p>}
            {loading ? (
                <p className="u-text-dim">Cargando servicios...</p>
            ) : (
                <>
                    {bundles.length > 0 && (
                        <div className="catalogSection">
                            <h4 className="catalogSectionTitle">Paquetes / Bundles</h4>
                            {renderServiceList(bundles)}
                        </div>
                    )}
                    <div className="catalogSection">
                        <h4 className="catalogSectionTitle">Servicios individuales</h4>
                        {standaloneServices.length > 0 ? (
                            renderServiceList(standaloneServices)
                        ) : (
                            <p className="u-text-sm u-text-dim">No hay servicios disponibles</p>
                        )}
                    </div>
                </>
            )}
        </>
    )

    return (
        <>
            {/* Desktop: modal overlay */}
            {open && (
                <div className="modalOverlay" onClick={onClose} role="presentation">
                    <div
                        className="modalContent catalogModal"
                        onClick={(event) => event.stopPropagation()}
                        role="dialog"
                        aria-modal="true"
                        aria-label="Asociar servicios"
                    >
                        <div className="modalHeader">
                            <h3 className="modalTitle">Asociar servicios</h3>
                            <button type="button" className="btn btnGhost btnSm" onClick={onClose}>
                                ✕
                            </button>
                        </div>
                        <div className="modalBody">
                            {content}
                        </div>
                        <div className="modalFooter">
                            <button type="button" className="btn btnPrimary" onClick={onClose}>
                                Listo
                            </button>
                        </div>
                    </div>
                </div>
            )}

            {/* Mobile: BottomSheet */}
            <BottomSheet open={open} title="Asociar servicios" onClose={onClose}>
                {content}
            </BottomSheet>
        </>
    )
}
