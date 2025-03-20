import { getAuthMeOptions, postAuthLogoutMutation } from '@/api/client/@tanstack/react-query.gen'
import { AuthUserResponse } from '@/api/client/types.gen'
import { useMutation, useQuery } from '@tanstack/react-query'
import { createContext, ReactNode, useContext } from 'react'

interface AuthContextType {
	user: AuthUserResponse | null
	isLoading: boolean
	logout: () => Promise<void>
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

export function AuthProvider({ children }: { children: ReactNode }) {
	const { data: user, isLoading: userLoading } = useQuery({
		...getAuthMeOptions({}),
		retry: false,
		retryDelay: 100,
	})

	const { mutateAsync: logout } = useMutation({
		...postAuthLogoutMutation({}),
		onSuccess: () => {
			location.reload()
		},
	})
	const value = {
		user: user || null,
		isLoading: userLoading,
		logout: async () => {
			await logout({})
		},
	}

	return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth() {
	const context = useContext(AuthContext)
	if (context === undefined) {
		throw new Error('useAuth must be used within an AuthProvider')
	}
	return context
}
