import { getNotificationsProvidersOptions, postNotificationsProvidersByIdTestMutation } from '@/api/client/@tanstack/react-query.gen'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuSeparator, DropdownMenuTrigger } from '@/components/ui/dropdown-menu'
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { Skeleton } from '@/components/ui/skeleton'
import { useMutation, useQuery } from '@tanstack/react-query'
import { formatDistanceToNow } from 'date-fns'
import { EllipsisVertical, Loader2, Mail } from 'lucide-react'
import { useNavigate } from 'react-router-dom'
import { toast } from 'sonner'
import { useForm } from 'react-hook-form'
import { z } from 'zod'
import { zodResolver } from '@hookform/resolvers/zod'
import { useState } from 'react'

const testEmailSchema = z.object({
	email: z.string().email('Please enter a valid email address'),
})

type TestEmailFormValues = z.infer<typeof testEmailSchema>

export function NotificationProvidersList() {
	const navigate = useNavigate()
	const [testProviderId, setTestProviderId] = useState<number | null>(null)

	const form = useForm<TestEmailFormValues>({
		resolver: zodResolver(testEmailSchema),
		defaultValues: {
			email: '',
		},
	})

	const {
		data: providers,
		isLoading,
		refetch,
	} = useQuery({
		...getNotificationsProvidersOptions({
			throwOnError: true,
			cache: 'no-cache',
		}),
	})

	const testMutation = useMutation({
		...postNotificationsProvidersByIdTestMutation({
			body: {
				testEmail: form.getValues('email'),
			},
		}),
		onSuccess: () => {
			toast.success('Provider tested successfully')
			refetch()
			setTestProviderId(null)
			form.reset()
		},
		onError: (error) => {
			toast.error('Failed to send test email', {
				description: error.message,
			})
			refetch()
			setTestProviderId(null)
		},
	})

	const onTestSubmit = (values: TestEmailFormValues) => {
		if (!testProviderId) return

		testMutation.mutate({
			path: { id: testProviderId },
			body: { testEmail: values.email },
		})
	}

	if (isLoading) {
		return <Skeleton className="h-48" />
	}

	return (
		<>
			<div className="space-y-4">
				<div className="flex justify-between">
					<h2 className="text-lg font-semibold">Notification Providers</h2>
					<Button onClick={() => navigate('/monitoring/providers/new')}>Add Provider</Button>
				</div>

				{!providers?.length ? (
					<Card className="flex h-[180px] flex-col items-center justify-center text-center">
						<CardContent className="pt-6">
							<div className="mb-4 flex justify-center">
								<Mail className="h-8 w-8 text-muted-foreground" />
							</div>
							<CardTitle className="text-lg font-semibold">No providers configured</CardTitle>
							<CardDescription className="mt-2">Get started by adding your first notification provider.</CardDescription>
							<Button onClick={() => navigate('/monitoring/providers/new')} className="mt-4">
								Add Provider
							</Button>
						</CardContent>
					</Card>
				) : (
					<div className="grid gap-4">
						{providers.map((provider) => (
							<Card key={provider.id}>
								<CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
									<div className="flex items-center gap-2">
										<div>
											<CardTitle>{provider.name}</CardTitle>
											<CardDescription>{provider.type}</CardDescription>
										</div>
										{provider.isDefault && <Badge variant="secondary">Default</Badge>}
									</div>
									<div className="flex items-center gap-2">
										<Button
											variant="outline"
											size="sm"
											onClick={() => setTestProviderId(provider.id)}
											disabled={testMutation.isPending && testMutation.variables?.path.id === provider.id}
										>
											{testMutation.isPending && testMutation.variables?.path.id === provider.id ? (
												<>
													<Loader2 className="mr-2 h-4 w-4 animate-spin" />
													Testing...
												</>
											) : (
												'Test'
											)}
										</Button>
										<DropdownMenu>
											<DropdownMenuTrigger asChild>
												<Button variant="ghost" size="icon">
													<EllipsisVertical className="h-4 w-4" />
												</Button>
											</DropdownMenuTrigger>
											<DropdownMenuContent align="end">
												<DropdownMenuItem onClick={() => navigate(`/monitoring/providers/${provider.id}`)}>Edit</DropdownMenuItem>
												<DropdownMenuSeparator />
												<DropdownMenuItem className="text-destructive">Delete</DropdownMenuItem>
											</DropdownMenuContent>
										</DropdownMenu>
									</div>
								</CardHeader>
								<CardContent>
									<div className="grid gap-2">
										<div className="text-sm">
											<span className="font-medium">Host:</span> {(provider.config as any).host}
										</div>
										<div className="text-sm">
											<span className="font-medium">From:</span> {(provider.config as any).from}
										</div>
										{provider.isDefault && <div className="text-sm text-muted-foreground">Default Provider</div>}
										<div className="text-sm text-muted-foreground">
											{provider.lastTestAt ? (
												<>
													Last tested {formatDistanceToNow(new Date(provider.lastTestAt), { addSuffix: true })}
													{provider.lastTestStatus && (
														<>
															{' • '}
															<span className={provider.lastTestStatus === 'success' ? 'text-green-500' : 'text-destructive'}>
																{provider.lastTestStatus === 'success' ? 'Success' : 'Failed'}
															</span>
														</>
													)}
													{provider.lastTestMessage && (
														<>
															{' • '}
															<span className="text-muted-foreground">{provider.lastTestMessage}</span>
														</>
													)}
												</>
											) : (
												'Never tested'
											)}
										</div>
									</div>
								</CardContent>
							</Card>
						))}
					</div>
				)}
			</div>

			<Dialog open={testProviderId !== null} onOpenChange={() => setTestProviderId(null)}>
				<DialogContent>
					<DialogHeader>
						<DialogTitle>Test Email Provider</DialogTitle>
						<DialogDescription>Enter the email address where you'd like to receive the test message.</DialogDescription>
					</DialogHeader>

					<Form {...form}>
						<form onSubmit={form.handleSubmit(onTestSubmit)} className="space-y-4">
							<FormField
								control={form.control}
								name="email"
								render={({ field }) => (
									<FormItem>
										<FormLabel>Email Address</FormLabel>
										<FormControl>
											<Input {...field} placeholder="test@example.com" type="email" />
										</FormControl>
										<FormMessage />
									</FormItem>
								)}
							/>

							<DialogFooter>
								<Button variant="outline" type="button" onClick={() => setTestProviderId(null)}>
									Cancel
								</Button>
								<Button type="submit" disabled={testMutation.isPending}>
									{testMutation.isPending ? (
										<>
											<Loader2 className="mr-2 h-4 w-4 animate-spin" />
											Sending...
										</>
									) : (
										'Send Test Email'
									)}
								</Button>
							</DialogFooter>
						</form>
					</Form>
				</DialogContent>
			</Dialog>
		</>
	)
}
