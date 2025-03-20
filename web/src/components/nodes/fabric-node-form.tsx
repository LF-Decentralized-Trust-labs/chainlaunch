import { GetNodesDefaultsFabricPeerResponse, GetNodesDefaultsFabricOrdererResponse } from '@/api/client'
import { Button } from '@/components/ui/button'
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage, FormDescription } from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { zodResolver } from '@hookform/resolvers/zod'
import { useEffect } from 'react'
import { useForm } from 'react-hook-form'
import { toast } from 'sonner'
import * as z from 'zod'

const fabricNodeFormSchema = z.object({
	name: z.string().min(2, 'Name must be at least 2 characters'),
	fabricProperties: z.object({
		nodeType: z.enum(['FABRIC_PEER', 'FABRIC_ORDERER']),
		mode: z.enum(['docker', 'service'], {
			required_error: 'Mode is required',
			invalid_type_error: 'Mode must be either docker or service',
		}),
		version: z.enum(['2.5.12', '3.0.0'], {
			required_error: 'Version is required',
			invalid_type_error: 'Invalid version selected',
		}),
		organizationId: z.number().min(1, 'Organization is required'),
		listenAddress: z.string().min(1, 'Listen address is required'),
		operationsListenAddress: z.string().min(1, 'Operations listen address is required'),
		externalEndpoint: z.string().optional(),
		domains: z.array(z.string()).optional(),
		// Peer-specific fields
		chaincodeAddress: z.string().optional(),
		eventsAddress: z.string().optional(),
		// Orderer-specific fields
		adminAddress: z.string().optional(),
	}),
})

export type FabricNodeFormValues = z.infer<typeof fabricNodeFormSchema>

interface FabricNodeFormProps {
	onSubmit: (data: FabricNodeFormValues) => void
	isSubmitting?: boolean
	organizations?: { id: number; name: string }[]
	defaults?: GetNodesDefaultsFabricPeerResponse | GetNodesDefaultsFabricOrdererResponse
	onNodeTypeChange?: (type: 'FABRIC_PEER' | 'FABRIC_ORDERER') => void
	hideSubmit?: boolean
	hideOrganization?: boolean
	hideNodeType?: boolean
	defaultValues?: FabricNodeFormValues
}

export function FabricNodeForm({ onSubmit, isSubmitting, organizations, defaults, onNodeTypeChange, hideSubmit, hideOrganization, hideNodeType, defaultValues }: FabricNodeFormProps) {
	const form = useForm<FabricNodeFormValues>({
		resolver: zodResolver(fabricNodeFormSchema),
		defaultValues: defaultValues || {
			name: '',
			fabricProperties: {
				nodeType: 'FABRIC_PEER',
				mode: 'service',
				version: '2.5.12',
				organizationId: undefined,
				listenAddress: defaults?.listenAddress || '',
				operationsListenAddress: defaults?.operationsListenAddress || '',
				externalEndpoint: defaults?.externalEndpoint || '',
				domains: [],
			},
		},
	})

	const nodeType = form.watch('fabricProperties.nodeType')

	useEffect(() => {
		if (defaults) {
			form.setValue('fabricProperties.listenAddress', defaults.listenAddress || '')
			form.setValue('fabricProperties.operationsListenAddress', defaults.operationsListenAddress || '')
			form.setValue('fabricProperties.externalEndpoint', defaults.externalEndpoint || '')

			if (nodeType === 'FABRIC_PEER' && 'chaincodeAddress' in defaults) {
				form.setValue('fabricProperties.chaincodeAddress', defaults.chaincodeAddress || '')
				form.setValue('fabricProperties.eventsAddress', defaults.eventsAddress || '')
			} else if (nodeType === 'FABRIC_ORDERER' && 'adminAddress' in defaults) {
				form.setValue('fabricProperties.adminAddress', defaults.adminAddress || '')
			}
		}
	}, [defaults, form, nodeType])

	const handleNodeTypeChange = (value: 'FABRIC_PEER' | 'FABRIC_ORDERER') => {
		form.setValue('fabricProperties.nodeType', value)
		// Clear type-specific fields when switching
		if (value === 'FABRIC_PEER') {
			form.setValue('fabricProperties.adminAddress', undefined)
		} else {
			form.setValue('fabricProperties.chaincodeAddress', undefined)
			form.setValue('fabricProperties.eventsAddress', undefined)
		}
		onNodeTypeChange?.(value)
	}

	return (
		<Form {...form}>
			<form
				onSubmit={form.handleSubmit(onSubmit, (errors) => {
					console.log(errors)
					toast.error(
						`Errors in the form: ${Object.values(errors)
							.map((error) => error.message)
							.join(', ')}`
					)
				})}
				className="space-y-6"
			>
				<FormField
					control={form.control}
					name="name"
					render={({ field }) => (
						<FormItem>
							<FormLabel>Node Name</FormLabel>
							<FormControl>
								<Input placeholder="Enter node name" {...field} />
							</FormControl>
							<FormMessage />
						</FormItem>
					)}
				/>

				{!hideOrganization && (
					<FormField
						control={form.control}
						name="fabricProperties.organizationId"
						render={({ field }) => (
							<FormItem>
								<FormLabel>Organization</FormLabel>
								<Select onValueChange={(value) => field.onChange(parseInt(value))} value={field.value?.toString()}>
									<FormControl>
										<SelectTrigger>
											<SelectValue placeholder="Select organization" />
										</SelectTrigger>
									</FormControl>
									<SelectContent>
										{organizations?.map((org) => (
											<SelectItem key={org.id} value={org.id.toString()}>
												{org.name}
											</SelectItem>
										))}
									</SelectContent>
								</Select>
								<FormMessage />
							</FormItem>
						)}
					/>
				)}

				<div className="grid gap-4 md:grid-cols-2">
					{!hideNodeType && (
						<FormField
							control={form.control}
							name="fabricProperties.nodeType"
							render={({ field }) => (
								<FormItem>
									<FormLabel>Node Type</FormLabel>
									<Select onValueChange={handleNodeTypeChange} value={field.value}>
										<FormControl>
											<SelectTrigger>
												<SelectValue placeholder="Select node type" />
											</SelectTrigger>
										</FormControl>
										<SelectContent>
											<SelectItem value="FABRIC_PEER">Peer</SelectItem>
											<SelectItem value="FABRIC_ORDERER">Orderer</SelectItem>
										</SelectContent>
									</Select>
									<FormMessage />
								</FormItem>
							)}
						/>
					)}

					<FormField
						control={form.control}
						name="fabricProperties.version"
						render={({ field }) => (
							<FormItem>
								<FormLabel>Version</FormLabel>
								<Select onValueChange={field.onChange} value={field.value}>
									<FormControl>
										<SelectTrigger>
											<SelectValue placeholder="Select version" />
										</SelectTrigger>
									</FormControl>
									<SelectContent>
										<SelectItem value="2.5.12">2.5.12</SelectItem>
										<SelectItem value="3.0.0">3.0.0</SelectItem>
									</SelectContent>
								</Select>
								<FormDescription>Select the Fabric version to use</FormDescription>
								<FormMessage />
							</FormItem>
						)}
					/>

					<FormField
						control={form.control}
						name="fabricProperties.mode"
						render={({ field }) => (
							<FormItem>
								<FormLabel>Mode</FormLabel>
								<Select onValueChange={field.onChange} value={field.value}>
									<FormControl>
										<SelectTrigger>
											<SelectValue placeholder="Select deployment mode" />
										</SelectTrigger>
									</FormControl>
									<SelectContent>
										<SelectItem value="docker">Docker</SelectItem>
										<SelectItem value="service">Service</SelectItem>
									</SelectContent>
								</Select>
								<FormDescription>Choose how the node will be deployed</FormDescription>
								<FormMessage />
							</FormItem>
						)}
					/>
				</div>

				<div className="space-y-4">
					<h3 className="text-lg font-medium">Network Configuration</h3>
					<div className="grid gap-4 md:grid-cols-2">
						<FormField
							control={form.control}
							name="fabricProperties.listenAddress"
							render={({ field }) => (
								<FormItem>
									<FormLabel>Listen Address</FormLabel>
									<FormControl>
										<Input placeholder="e.g., 0.0.0.0:7051" {...field} />
									</FormControl>
									<FormMessage />
								</FormItem>
							)}
						/>

						<FormField
							control={form.control}
							name="fabricProperties.operationsListenAddress"
							render={({ field }) => (
								<FormItem>
									<FormLabel>Operations Address</FormLabel>
									<FormControl>
										<Input placeholder="e.g., 0.0.0.0:9443" {...field} />
									</FormControl>
									<FormMessage />
								</FormItem>
							)}
						/>

						<FormField
							control={form.control}
							name="fabricProperties.externalEndpoint"
							render={({ field }) => (
								<FormItem>
									<FormLabel>External Endpoint</FormLabel>
									<FormControl>
										<Input placeholder="e.g., peer0.org1.example.com:7051" {...field} />
									</FormControl>
									<FormMessage />
								</FormItem>
							)}
						/>
					</div>
				</div>

				<div className="space-y-4">
					<h3 className="text-lg font-medium">Domain Configuration</h3>
					<div className="grid gap-4">
						<FormField
							control={form.control}
							name="fabricProperties.domains"
							render={({ field }) => (
								<FormItem>
									<FormLabel>Domains</FormLabel>
									<FormDescription>Enter additional domains or IP addresses (one per line). Note: localhost and 127.0.0.1 are added by default</FormDescription>
									<FormControl>
										<textarea
											className="min-h-[100px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
											placeholder="example.com&#10;192.168.1.1"
											onChange={(e) => {
												const domains = e.target.value
													.split('\n')
													.map((d) => d.trim())
													.filter(Boolean)
												field.onChange(domains)
											}}
											value={field.value?.join('\n') || ''}
										/>
									</FormControl>
									<FormMessage />
								</FormItem>
							)}
						/>
					</div>
				</div>

				{nodeType === 'FABRIC_PEER' && (
					<div className="space-y-4">
						<h3 className="text-lg font-medium">Peer Configuration</h3>
						<div className="grid gap-4 md:grid-cols-2">
							<FormField
								control={form.control}
								name="fabricProperties.chaincodeAddress"
								render={({ field }) => (
									<FormItem>
										<FormLabel>Chaincode Address</FormLabel>
										<FormControl>
											<Input placeholder="e.g., 0.0.0.0:7052" {...field} />
										</FormControl>
										<FormMessage />
									</FormItem>
								)}
							/>

							<FormField
								control={form.control}
								name="fabricProperties.eventsAddress"
								render={({ field }) => (
									<FormItem>
										<FormLabel>Events Address</FormLabel>
										<FormControl>
											<Input placeholder="e.g., 0.0.0.0:7053" {...field} />
										</FormControl>
										<FormMessage />
									</FormItem>
								)}
							/>
						</div>
					</div>
				)}

				{nodeType === 'FABRIC_ORDERER' && (
					<div className="space-y-4">
						<h3 className="text-lg font-medium">Orderer Configuration</h3>
						<div className="grid gap-4 md:grid-cols-2">
							<FormField
								control={form.control}
								name="fabricProperties.adminAddress"
								render={({ field }) => (
									<FormItem>
										<FormLabel>Admin Address</FormLabel>
										<FormControl>
											<Input placeholder="e.g., 0.0.0.0:7053" {...field} />
										</FormControl>
										<FormMessage />
									</FormItem>
								)}
							/>
						</div>
					</div>
				)}

				{!hideSubmit && (
					<Button type="submit" disabled={isSubmitting}>
						{isSubmitting ? 'Creating...' : 'Create Node'}
					</Button>
				)}
			</form>
		</Form>
	)
}
