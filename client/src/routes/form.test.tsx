import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { SimpleFormExample } from './form'

describe('Form Component', () => {
  beforeEach(() => {
    // Clear console.log calls between tests
    vi.spyOn(console, 'log').mockImplementation(() => {})
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('renders form with initial state', () => {
    render(<SimpleFormExample />)

    expect(screen.getByText('Simple Form Example')).toBeInTheDocument()
    expect(screen.getByLabelText('First Name')).toBeInTheDocument()
    expect(screen.getByLabelText('Last Name')).toBeInTheDocument()
    // TanStack Form enables submit when no validation errors exist
    expect(screen.getByRole('button', { name: 'Submit' })).not.toBeDisabled()
  })

  it('validates firstName field correctly', async () => {
    render(<SimpleFormExample />)

    const firstNameInput = screen.getByLabelText('First Name')

    // Test minimum length validation - first enter a short value to trigger touched state
    fireEvent.change(firstNameInput, { target: { value: 'Jo' } })
    fireEvent.blur(firstNameInput)

    await waitFor(() => {
      expect(screen.getByText('First name must be at least 3 characters')).toBeInTheDocument()
    })

    // Test empty field validation
    fireEvent.change(firstNameInput, { target: { value: '' } })
    fireEvent.blur(firstNameInput)

    await waitFor(() => {
      expect(screen.getByText('A first name is required')).toBeInTheDocument()
    })

    // Test valid input clears error
    fireEvent.change(firstNameInput, { target: { value: 'John' } })
    fireEvent.blur(firstNameInput)

    await waitFor(() => {
      expect(screen.queryByText('A first name is required')).not.toBeInTheDocument()
      expect(screen.queryByText('First name must be at least 3 characters')).not.toBeInTheDocument()
    })
  })

  it('shows async validation error for "error" in firstName', async () => {
    render(<SimpleFormExample />)

    const firstNameInput = screen.getByLabelText('First Name')

    // Type "error" - this will trigger async validation with debounce
    fireEvent.change(firstNameInput, { target: { value: 'error' } })

    // Wait for debounce (500ms) + async validation (1000ms) + some buffer
    await waitFor(() => {
      expect(screen.getByText('No "error" allowed in first name')).toBeInTheDocument()
    }, { timeout: 3000 })
  })

  it('disables submit button when form has validation errors', async () => {
    render(<SimpleFormExample />)

    const firstNameInput = screen.getByLabelText('First Name')
    const submitButton = screen.getByRole('button', { name: 'Submit' })

    // Initially enabled (no validation errors)
    expect(submitButton).not.toBeDisabled()

    // Enter invalid data to trigger validation error
    fireEvent.change(firstNameInput, { target: { value: 'Jo' } }) // Too short
    fireEvent.blur(firstNameInput)

    // Wait for validation to complete and button to be disabled
    await waitFor(() => {
      expect(submitButton).toBeDisabled()
    })

    // Fix the validation error
    fireEvent.change(firstNameInput, { target: { value: 'John' } })
    fireEvent.blur(firstNameInput)

    // Button should be enabled again
    await waitFor(() => {
      expect(submitButton).not.toBeDisabled()
    })
  })

  it('renders submit button correctly', () => {
    render(<SimpleFormExample />)

    const submitButton = screen.getByRole('button', { name: 'Submit' })

    // Button should exist and be enabled initially
    expect(submitButton).toBeInTheDocument()
    expect(submitButton).not.toBeDisabled()
  })
})