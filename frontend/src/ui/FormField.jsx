/**
 * FormField — labeled input wrapper.
 * Ensures consistent label–input spacing and font throughout the app.
 *
 * @param {string} label     - Visible label text
 * @param {string} [htmlFor] - Ties label to input by id (optional; usually set by parent)
 * @param {React.ReactNode} children - The input element(s)
 */
export function FormField({ label, htmlFor, children }) {
  return (
    <div className="formField">
      <label className="formLabel" htmlFor={htmlFor}>
        {label}
      </label>
      {children}
    </div>
  )
}
