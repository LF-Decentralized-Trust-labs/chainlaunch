import { useAuth } from '@/contexts/AuthContext'
import LoginPage from '@/pages/Login'

export const ProtectedLayout = ({ children }: { children: React.ReactNode }) => {
	const { user, isLoading } = useAuth()

	if (isLoading) {
		return <div className="flex items-center justify-center h-screen">Loading...</div>
	}

	if (!user) {
		return <LoginPage />
	}

	return <>{children}</>
}
