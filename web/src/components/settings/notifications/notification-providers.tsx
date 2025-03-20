import { postNotificationsProviders } from '@/api/client'
import { getNotificationsProvidersOptions } from '@/api/client/@tanstack/react-query.gen'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { Skeleton } from '@/components/ui/skeleton'
import { zodResolver } from '@hookform/resolvers/zod'
import { useMutation, useQuery } from '@tanstack/react-query'
import { useForm } from 'react-hook-form'
import { toast } from 'sonner'
import { z } from 'zod'

const providerFormSchema = z.object({
	name: z.string().min(1, 'Name is required'),
	type: z.literal('SMTP'),
	isDefault: z.boolean().optional(),
	config: z.object({
		host: z.string().min(1, 'SMTP host is required'),
		port: z.coerce.number().min(1, 'SMTP port is required'),
		username: z.string().min(1, 'SMTP username is required'),
		password: z.string().min(1, 'SMTP password is required'),
		from: z.string().email('Must be a valid email address'),
	}),
})

type ProviderFormValues = z.infer<typeof providerFormSchema>

export function NotificationProviders() {
	// Fetch providers
	const { data: providers, isLoading } = useQuery({
		...getNotificationsProvidersOptions({ throwOnError: true }),
	})

	const form = useForm<ProviderFormValues>({
		resolver: zodResolver(providerFormSchema),
		defaultValues: {
			type: 'SMTP',
			isDefault: false,
			config: {
				host: '',
				port: 587,
				username: '',
				password: '',
				from: '',
			},
		},
	})

	// Create provider
	const createMutation = useMutation({
		mutationFn: (values: ProviderFormValues) => {
			return postNotificationsProviders({
				throwOnError: true,
				body: {
					name: values.name,
					type: values.type,
					isDefault: values.isDefault,
					config: values.config,
				},
			})
		},
		onSuccess: () => {
			toast.success('Email provider created successfully')
			form.reset()
		},
		onError: (error: Error) => {
			toast.error('Failed to create email provider', {
				description: error.message,
			})
		},
	})

	if (isLoading) {
		return <Skeleton className="h-48" />
	}

	return (
		<Card>
			<CardHeader>
				<CardTitle>Email Provider</CardTitle>
				<CardDescription>Configure your SMTP email provider for notifications</CardDescription>
			</CardHeader>
			<CardContent>
				<Form {...form}>
					<form onSubmit={form.handleSubmit((data) => createMutation.mutate(data))} className="space-y-4">
						<FormField
							control={form.control}
							name="name"
							render={({ field }) => (
								<FormItem>
									<FormLabel>Provider Name</FormLabel>
									<FormControl>
										<Input {...field} placeholder="My SMTP Provider" />
									</FormControl>
									<FormMessage />
								</FormItem>
							)}
						/>

						<div className="grid gap-4 md:grid-cols-2">
							<FormField
								control={form.control}
								name="config.host"
								render={({ field }) => (
									<FormItem>
										<FormLabel>SMTP Host</FormLabel>
										<FormControl>
											<Input {...field} placeholder="smtp.example.com" />
										</FormControl>
										<FormMessage />
									</FormItem>
								)}
							/>

							<FormField
								control={form.control}
								name="config.port"
								render={({ field }) => (
									<FormItem>
										<FormLabel>SMTP Port</FormLabel>
										<FormControl>
											<Input {...field} type="number" />
										</FormControl>
										<FormMessage />
									</FormItem>
								)}
							/>

							<FormField
								control={form.control}
								name="config.username"
								render={({ field }) => (
									<FormItem>
										<FormLabel>SMTP Username</FormLabel>
										<FormControl>
											<Input {...field} />
										</FormControl>
										<FormMessage />
									</FormItem>
								)}
							/>

							<FormField
								control={form.control}
								name="config.password"
								render={({ field }) => (
									<FormItem>
										<FormLabel>SMTP Password</FormLabel>
										<FormControl>
											<Input {...field} type="password" />
										</FormControl>
										<FormMessage />
									</FormItem>
								)}
							/>

							<FormField
								control={form.control}
								name="config.from"
								render={({ field }) => (
									<FormItem>
										<FormLabel>From Address</FormLabel>
										<FormControl>
											<Input {...field} placeholder="notifications@example.com" />
										</FormControl>
										<FormMessage />
									</FormItem>
								)}
							/>
						</div>

						<Button type="submit" disabled={createMutation.isPending}>
							Create Provider
						</Button>
					</form>
				</Form>
			</CardContent>
		</Card>
	)
}
