import { LoginForm } from '@/components/auth/login-form'
import { useMutation } from '@tanstack/react-query'
import { postAuthLoginMutation } from '@/api/client/@tanstack/react-query.gen'
import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { toast } from 'sonner'

export default function LoginPage() {
	const [isLoading, setIsLoading] = useState(false)
	const navigate = useNavigate()
	const loginMutation = useMutation({
		...postAuthLoginMutation(),
		onSuccess: () => {
			navigate('/nodes')
			toast.success('Logged in successfully')
			location.reload()
		},
	})
	const handleSubmit = async (data: { username: string; password: string }) => {
		setIsLoading(true)
		try {
			await loginMutation.mutateAsync({
				body: data,
			})
			// Store auth token or user data in localStorage/state management
			localStorage.setItem('isAuthenticated', 'true')

			toast.success('Logged in successfully')
			navigate('/nodes')
		} catch (error) {
			toast.error('Invalid credentials')
		} finally {
			setIsLoading(false)
		}
	}

	return (
		<div className="min-h-screen flex items-center justify-center p-4 bg-background">
			<LoginForm onSubmit={handleSubmit} isLoading={isLoading} />
		</div>
	)
}
