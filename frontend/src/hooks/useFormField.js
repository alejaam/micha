import { useState } from 'react'

/**
 * useFormField — lightweight form field state manager.
 *
 * @param {string} initialValue
 * @param {(value: string) => string | null} [validator] — returns error string or null
 * @returns {{ value: string, error: string | null, touched: boolean, setValue: (v: string) => void, onBlur: () => void, setTouched: (v: boolean) => void, isValid: boolean }}
 */
export function useFormField(initialValue = '', validator) {
    const [value, setValue] = useState(initialValue)
    const [touched, setTouched] = useState(false)

    const error = touched && validator ? validator(value) : null
    const isValid = !error

    const onBlur = () => setTouched(true)

    return { value, error, touched, setValue, onBlur, setTouched, isValid }
}
