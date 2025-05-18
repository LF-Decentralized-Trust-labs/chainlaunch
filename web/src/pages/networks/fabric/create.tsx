import { HandlerOrganizationResponse, HttpNodeResponse } from '@/api/client'
import { getNodesOptions, getOrganizationsOptions, postNetworksFabricMutation } from '@/api/client/@tanstack/react-query.gen'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Checkbox } from '@/components/ui/checkbox'
import { Form, FormControl, FormDescription, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { Separator } from '@/components/ui/separator'
import { Skeleton } from '@/components/ui/skeleton'
import { Switch } from '@/components/ui/switch'
import { cn } from '@/lib/utils'
import { zodResolver } from '@hookform/resolvers/zod'
import { useMutation, useQuery } from '@tanstack/react-query'
import { AlertCircle, TriangleAlert } from 'lucide-react'
import { useEffect, useMemo, useState } from 'react'
import { useForm } from 'react-hook-form'
import { useNavigate } from 'react-router-dom'
import { toast } from 'sonner'
import * as z from 'zod'

interface OrganizationWithNodes extends HandlerOrganizationResponse {
	orderers: HttpNodeResponse[]
}

const channelFormSchema = z
	.object({
		name: z.string().min(1, 'Channel name is required'),
		organizations: z.array(
			z.object({
				id: z.number(),
				enabled: z.boolean().default(false),
				isPeer: z.boolean().default(false),
				isOrderer: z.boolean().default(false),
				consenters: z.array(z.number()),
			})
		),
	})
	.refine(
		(data) => {
			const enabledOrgs = [...data.organizations.filter((org) => org.enabled)]

			// At least one peer organization must be enabled
			return enabledOrgs.some((org) => org.isPeer)
		},
		{ message: 'At least one peer organization must be enabled' }
	)
	.refine(
		(data) => {
			const enabledOrdererOrgs = data.organizations.filter((org) => org.enabled && org.isOrderer)

			// At least one orderer organization must have consenters
			return enabledOrdererOrgs.some((org) => org.consenters.length > 0)
		},
		{ message: 'At least one orderer organization with consenters is required' }
	)

type ChannelFormValues = z.infer<typeof channelFormSchema>

export default function FabricCreateChannel() {
	const { data: organizations, isLoading: isLoadingOrgs } = useQuery({
		...getOrganizationsOptions(),
	})
	const { data: nodes, isLoading: isLoadingNodes } = useQuery({
		...getNodesOptions({
			query: {
				limit: 1000,
				page: 1,
				platform: 'FABRIC',
			},
		}),
	})

	const navigate = useNavigate()
	const createNetwork = useMutation({
		...postNetworksFabricMutation(),
		onSuccess: (network) => {
			toast.success('Network created successfully')
			navigate(`/networks/${network.id}/fabric`)
		},
		onError: (error: any) => {
			toast.error(`Failed to create network: ${error.message}`)
		},
	})
	const form = useForm<ChannelFormValues>({
		resolver: zodResolver(channelFormSchema),
		defaultValues: {
			name: '',
			organizations: [],
		},
	})

	const [formError, setFormError] = useState<string | null>(null)

	// Update form values when queries complete for local organizations
	useEffect(() => {
		if (organizations && nodes?.items) {
			const defaultOrgs = organizations.items?.map((org) => {
				const orderers = nodes.items?.filter((node) => node.platform === 'FABRIC' && node.nodeType === 'FABRIC_ORDERER' && node.fabricOrderer?.mspId === org.mspId)

				return {
					id: org.id!,
					enabled: true,
					isPeer: true,
					isOrderer: true,
					consenters: orderers?.map((orderer) => orderer.id!) || [],
				}
			})
			form.setValue('organizations', defaultOrgs, { shouldDirty: true })
		}
	}, [organizations, nodes, form])

	// Process external organizations and orderers

	const organizationsWithNodes = useMemo(
		() =>
			organizations?.items?.map((org) => ({
				...org,
				orderers: nodes?.items?.filter((node) => node.platform === 'FABRIC' && node.nodeType === 'FABRIC_ORDERER' && node.fabricOrderer?.mspId === org.mspId) || [],
			})) as OrganizationWithNodes[],
		[organizations, nodes]
	)

	const onSubmit = async (data: ChannelFormValues) => {
		try {
			// The code below won't execute due to the return statement above
			const enabledLocalOrgs = data.organizations
				.filter((org) => org.enabled)
				.map((org) => ({
					id: org.id,
					nodeIds: org.isOrderer ? org.consenters : [],
					isPeer: org.isPeer,
					isOrderer: org.isOrderer,
				}))

			await createNetwork.mutate({
				body: {
					name: data.name,
					config: {
						peerOrganizations: enabledLocalOrgs
							.filter((org) => org.isPeer)
							.map((org) => ({
								id: org.id,
								nodeIds: [],
							})),
						ordererOrganizations: enabledLocalOrgs
							.filter((org) => org.isOrderer)
							.map((org) => ({
								id: org.id,
								nodeIds: org.nodeIds,
							})),
					},
					description: '',
				},
			})
		} catch (error: any) {
			const errorMessage = error.message || 'An unexpected error occurred. Please try again later.'
			setFormError(errorMessage)
			toast.error('Network creation failed', {
				description: errorMessage,
			})
		}
	}

	const isLoading = isLoadingOrgs || isLoadingNodes

	return (
		<div className="flex-1 p-8">
			<div className="max-w-4xl mx-auto">
				<div className="mb-8">
					<h1 className="text-2xl font-semibold">Configure Channel</h1>
					<p className="text-muted-foreground">Create a new Fabric channel</p>
				</div>

				{formError && (
					<Alert variant="destructive" className="mb-6">
						<AlertCircle className="h-4 w-4" />
						<AlertTitle>Validation Error</AlertTitle>
						<AlertDescription>
							<div className="space-y-2">
								{formError.split('\n').map(
									(line, index) =>
										line.trim() && (
											<div key={index} className="flex items-start gap-2">
												<span>•</span>
												<span>{line.trim()}</span>
											</div>
										)
								)}
							</div>
						</AlertDescription>
					</Alert>
				)}

				<Form {...form}>
					<form
						onSubmit={form.handleSubmit(onSubmit, (errors) => {
							console.error('Form validation errors:', errors)

							// Create a more specific error message based on the actual validation errors
							const errorMessages: string[] = []
							Object.entries(errors).forEach(([key, value]) => {
								errorMessages.push(`${key ? `${key}: ${value.message || 'Unknown error'}` : value.message || 'Unknown error'}`)
							})

							// Set the specific error message
							setFormError(errorMessages.join('\n'))

							toast.error('Please fix the validation errors', {
								description: 'There are errors in the form that need to be fixed before you can create a network.',
							})
						})}
						className="space-y-8"
					>
						<Card className={cn(form.formState.errors.name && 'border-destructive')}>
							<CardHeader>
								<div className="flex items-center justify-between">
									<div>
										<CardTitle>Channel Information</CardTitle>
										<CardDescription>Basic channel configuration</CardDescription>
									</div>
									{form.formState.errors.name && <AlertCircle className="h-5 w-5 text-destructive" />}
								</div>
							</CardHeader>
							<CardContent>
								<FormField
									control={form.control}
									name="name"
									render={({ field }) => (
										<FormItem>
											<FormLabel>Channel Name</FormLabel>
											<FormControl>
												<Input placeholder="mychannel" {...field} />
											</FormControl>
											<FormDescription>The name of the channel to be created</FormDescription>
											<FormMessage />
										</FormItem>
									)}
								/>
							</CardContent>
						</Card>

						{form.formState.errors.organizations?.root?.message && (
							<div className="flex flex-col items-center justify-center p-8 border-2 border-dashed rounded-lg border-destructive/20 bg-destructive/5">
								<div className="h-12 w-12 text-destructive mb-4 flex items-center justify-center">
									<TriangleAlert className="h-8 w-8" />
								</div>
								<div className="space-y-2 text-center">
									<FormMessage>{form.formState.errors.organizations?.root?.message}</FormMessage>
									<p className="text-sm text-muted-foreground">
										Requirements:
										<ul className="list-disc list-inside mt-1">
											<li>At least one organization must be enabled</li>
											<li>At least one organization must have consenters</li>
											<li>Total consenters across all organizations must be at least 3</li>
										</ul>
									</p>
								</div>
							</div>
						)}

						<Card className={cn(form.formState.errors.organizations && 'border-destructive')}>
							<CardHeader>
								<div className="flex items-center justify-between">
									<div>
										<CardTitle>Local Organizations</CardTitle>
										<CardDescription>Configure organizations from your local network</CardDescription>
									</div>
									{form.formState.errors.organizations && <AlertCircle className="h-5 w-5 text-destructive" />}
								</div>
							</CardHeader>
							<CardContent className="space-y-6">
								{isLoadingOrgs || isLoadingNodes ? (
									Array.from({ length: 3 }).map((_, i) => (
										<div key={i} className="space-y-4 rounded-lg border p-4">
											<div className="flex items-center justify-between">
												<div className="flex items-center gap-4">
													<Skeleton className="h-4 w-[40px]" />
													<div>
														<Skeleton className="h-5 w-[150px] mb-2" />
														<Skeleton className="h-4 w-[200px]" />
													</div>
												</div>
												<Skeleton className="h-5 w-[100px]" />
											</div>
										</div>
									))
								) : organizationsWithNodes?.length === 0 ? (
									<div className="text-center p-4 border rounded-lg">
										<p className="text-muted-foreground">No local organizations available</p>
									</div>
								) : (
									organizationsWithNodes?.map((org, index) => (
										<div key={org.id} className={cn('space-y-4 rounded-lg border p-4', form.formState.errors.organizations?.[index] && 'border-destructive')}>
											<div className="flex items-center justify-between">
												<div className="flex items-center gap-4">
													<FormField
														control={form.control}
														name={`organizations.${index}.enabled`}
														render={({ field }) => (
															<FormItem className="flex items-center gap-2 space-y-0">
																<FormControl>
																	<Switch checked={field.value} onCheckedChange={field.onChange} />
																</FormControl>
																<div>
																	<h3 className="font-medium">{org.mspId}</h3>
																	<p className="text-sm text-muted-foreground">{org.description}</p>
																</div>
															</FormItem>
														)}
													/>
												</div>
												<Badge variant="outline">{org.orderers.length} Orderers</Badge>
											</div>

											{form.watch(`organizations.${index}.enabled`) && (
												<>
													<div className="flex gap-6">
														<FormField
															control={form.control}
															name={`organizations.${index}.isPeer`}
															render={({ field }) => (
																<FormItem className="flex items-center gap-2">
																	<FormControl>
																		<Checkbox checked={field.value} onCheckedChange={field.onChange} />
																	</FormControl>
																	<FormLabel className="!mt-0">Peer Organization</FormLabel>
																</FormItem>
															)}
														/>
														<FormField
															control={form.control}
															name={`organizations.${index}.isOrderer`}
															render={({ field }) => (
																<FormItem className="flex items-center gap-2">
																	<FormControl>
																		<Checkbox checked={field.value} onCheckedChange={field.onChange} />
																	</FormControl>
																	<FormLabel className="!mt-0">Orderer Organization</FormLabel>
																</FormItem>
															)}
														/>
													</div>

													{form.watch(`organizations.${index}.isOrderer`) && org.orderers.length > 0 && (
														<>
															<Separator />
															<FormField
																control={form.control}
																name={`organizations.${index}.consenters`}
																render={({ field }) => (
																	<FormItem>
																		<div className="flex items-center justify-between">
																			<FormLabel>Consenters</FormLabel>
																			{form.formState.errors.organizations?.[index]?.consenters && (
																				<div className="text-sm text-destructive">
																					{Array.isArray(form.formState.errors.organizations[index].consenters)
																						? form.formState.errors.organizations[index].consenters.map((err: any, i: number) => (
																								<span key={i}>{err?.message || JSON.stringify(err)}</span>
																						  ))
																						: form.formState.errors.organizations[index].consenters?.message ||
																						  JSON.stringify(form.formState.errors.organizations[index].consenters)}
																				</div>
																			)}
																		</div>
																		<div className="grid gap-2">
																			{org.orderers.map((orderer) => (
																				<FormItem key={orderer.id} className="flex items-center gap-2">
																					<FormControl>
																						<Checkbox
																							checked={field.value?.includes(orderer.id!)}
																							onCheckedChange={(checked) => {
																								const value = checked
																									? [...(field.value || []), orderer.id]
																									: (field.value || []).filter((id) => id !== orderer.id)
																								field.onChange(value)
																							}}
																						/>
																					</FormControl>
																					<FormLabel className="!mt-0">{orderer.name}</FormLabel>
																				</FormItem>
																			))}
																		</div>
																		{org.orderers.length > 0 && field.value?.length === 0 && form.watch(`organizations.${index}.isOrderer`) && (
																			<p className="text-sm text-amber-500 mt-2">
																				<AlertCircle className="h-3 w-3 inline-block mr-1" />
																				You should select at least one consenter for this orderer organization
																			</p>
																		)}
																		<FormMessage />
																	</FormItem>
																)}
															/>
														</>
													)}
												</>
											)}
										</div>
									))
								)}
							</CardContent>
						</Card>

						<div className="flex justify-end">
							<Button
								type="submit"
								disabled={form.formState.isSubmitting || isLoading}
								onClick={() => {
									// Check for specific validation errors when the button is clicked
									const errors = form.formState.errors
									if (!form.formState.isValid) {
										// Create a specific error message based on the validation errors
										let errorMessage = ''

										if (errors.name) {
											errorMessage += `Channel Name: ${errors.name.message}\n`
										}

										if (errors.organizations?.root) {
											errorMessage += `Organizations: ${errors.organizations.root.message}\n`
										} else if (errors.organizations) {
											// Check for errors in specific organizations
											const orgErrors = Object.entries(errors.organizations)
												.filter(([key, value]) => key !== 'root' && value)
												.map(([key, value]) => `Organization ${parseInt(key) + 1}: ${JSON.stringify(value)}`)

											if (orgErrors.length > 0) {
												errorMessage += `Organization errors:\n${orgErrors.join('\n')}\n`
											}
										}

										if (errorMessage) {
											setFormError(errorMessage)
											toast.error('Please fix the following validation errors:', {
												description: errorMessage,
											})
										}
									}
								}}
							>
								{form.formState.isSubmitting ? 'Creating...' : 'Create Channel'}
							</Button>
						</div>
					</form>
				</Form>
			</div>
		</div>
	)
}
