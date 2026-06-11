import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { RegisterPage } from '../RegisterPage'

const mockRegister = vi.fn()
const mockLogin = vi.fn()
const mockNavigate = vi.fn()

vi.mock('../../context/AuthContext', () => ({
    useAuth: () => ({
        register: mockRegister,
        login: mockLogin,
    }),
}))

vi.mock('react-router-dom', async () => {
    const actual = await vi.importActual('react-router-dom')
    return {
        ...actual,
        useNavigate: () => mockNavigate,
    }
})

describe('RegisterPage', () => {
    beforeEach(() => {
        mockRegister.mockReset()
        mockLogin.mockReset()
        mockNavigate.mockReset()
    })

    it('renders the name input field', () => {
        render(
            <MemoryRouter>
                <RegisterPage />
            </MemoryRouter>,
        )

        expect(screen.getByLabelText(/nombre completo/i)).toBeInTheDocument()
    })

    it('disables submit when name is empty', () => {
        render(
            <MemoryRouter>
                <RegisterPage />
            </MemoryRouter>,
        )

        const emailInput = screen.getByLabelText(/correo electrónico/i)
        const passwordInput = screen.getByLabelText(/^contraseña/i)
        const confirmPasswordInput = screen.getByLabelText(/confirmar contraseña/i)
        const nameInput = screen.getByLabelText(/nombre completo/i)
        const submitButton = screen.getByRole('button', { name: /crear cuenta/i })

        // Pre-fill valid email, password, and confirmation, but leave name empty
        fireEvent.change(emailInput, { target: { value: 'alice@example.com' } })
        fireEvent.change(passwordInput, { target: { value: 'password123' } })
        fireEvent.change(confirmPasswordInput, { target: { value: 'password123' } })
        fireEvent.change(nameInput, { target: { value: '' } })

        expect(submitButton).toBeDisabled()
    })

    it('submits name, email, and password on successful validation', async () => {
        mockRegister.mockResolvedValueOnce({})
        mockLogin.mockResolvedValueOnce({})

        render(
            <MemoryRouter>
                <RegisterPage />
            </MemoryRouter>,
        )

        const emailInput = screen.getByLabelText(/correo electrónico/i)
        const passwordInput = screen.getByLabelText(/^contraseña/i)
        const confirmPasswordInput = screen.getByLabelText(/confirmar contraseña/i)
        const nameInput = screen.getByLabelText(/nombre completo/i)
        const submitButton = screen.getByRole('button', { name: /crear cuenta/i })

        fireEvent.change(emailInput, { target: { value: 'alice@example.com' } })
        fireEvent.change(nameInput, { target: { value: 'Alice Smith' } })
        fireEvent.change(passwordInput, { target: { value: 'password123' } })
        fireEvent.change(confirmPasswordInput, { target: { value: 'password123' } })

        expect(submitButton).toBeEnabled()
        fireEvent.click(submitButton)

        await waitFor(() => {
            expect(mockRegister).toHaveBeenCalledWith({
                name: 'Alice Smith',
                email: 'alice@example.com',
                password: 'password123',
            })
        })
    })
})
