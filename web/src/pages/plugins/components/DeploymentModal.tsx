import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog'
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { zodResolver } from '@hookform/resolvers/zod'
import { useForm } from 'react-hook-form'
import { z } from 'zod'
import { useQuery } from '@tanstack/react-query'
import { getKeysOptions, getOrganizationsOptions, getNodesOptions } from '@/api/client/@tanstack/react-query.gen'
import { getKeysById } from '@/api/client/sdk.gen'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Checkbox } from '@/components/ui/checkbox'
import { ScrollArea } from '@/components/ui/scroll-area'
import { useCallback, useEffect, useMemo, useState } from 'react'
import { GetOrganizationsResponse, HandlerOrganizationResponse } from '@/api/client'

interface DeploymentModalProps {
	isOpen: boolean
	onClose: () => void
	onDeploy: (params: Record<string, unknown>) => void
	parameters?: Record<string, any> // JSON Schema
}

type FormValues = {
	[key: string]: number | number[] | string | boolean | Record<string, any>
}

interface FabricKeySelectProps {
	value?: number
	onChange: (value: number) => void
	disabled?: boolean
}

const FabricKeySelect = ({ value, onChange, disabled }: FabricKeySelectProps) => {
	const [selectedOrgId, setSelectedOrgId] = useState<number | null>(null)
	const [selectedKeys, setSelectedKeys] = useState<Array<{ id: number; name: string; description?: string; algorithm?: string; keySize?: number; curve?: string }>>([])
	const [isLoading, setIsLoading] = useState(false)

	const { data: organizations } = useQuery({
		...getOrganizationsOptions(),
	})
	// Get the selected organization
	const selectedOrg = useMemo(() => organizations?.items?.find((org) => org.id === selectedOrgId), [organizations, selectedOrgId])
	// Get key IDs for the selected organization
	const keyIds = useMemo(() => {
		if (!selectedOrg) return []
		const ids: number[] = []
		if (selectedOrg.adminSignKeyId) ids.push(selectedOrg.adminSignKeyId)
		if (selectedOrg.adminTlsKeyId) ids.push(selectedOrg.adminTlsKeyId)
		if (selectedOrg.clientSignKeyId) ids.push(selectedOrg.clientSignKeyId)
		return ids
	}, [selectedOrg])

	// Fetch key details when organization changes
	const fetchKeyDetails = useCallback(async () => {
		if (!keyIds.length) {
			setSelectedKeys([])
			return
		}

		setIsLoading(true)
		try {
			const keyDetails = await Promise.all(
				keyIds.map(async (keyId) => {
					const { data } = await getKeysById({ path: { id: keyId } })
					return {
						id: keyId,
						name: data.name || `Key ${keyId}`,
						description: data.description,
						algorithm: data.algorithm,
						keySize: data.keySize,
						curve: data.curve,
					}
				})
			)
			setSelectedKeys(keyDetails)
		} catch (error) {
			console.error('Error fetching key details:', error)
			setSelectedKeys([])
		} finally {
			setIsLoading(false)
		}
	}, [keyIds])

	useEffect(() => {
		if (selectedOrgId && keyIds.length > 0) {
			fetchKeyDetails()
		}
	}, [fetchKeyDetails, selectedOrgId, keyIds])

	return (
		<div className="space-y-4">
			<Select value={selectedOrgId?.toString()} onValueChange={(val) => setSelectedOrgId(Number(val))} disabled={disabled}>
				<SelectTrigger>
					<SelectValue placeholder="Select an organization" />
				</SelectTrigger>
				<SelectContent>
					<ScrollArea className="h-[200px]">
						{organizations?.items?.map((org) => (
							<SelectItem key={org.id} value={org.id?.toString()}>
								{org.mspId}
							</SelectItem>
						))}
					</ScrollArea>
				</SelectContent>
			</Select>

			<Select value={value?.toString()} onValueChange={(val) => onChange(Number(val))} disabled={disabled || !selectedOrgId || isLoading}>
				<SelectTrigger>
					<SelectValue placeholder={isLoading ? 'Loading keys...' : selectedOrgId ? 'Select a key' : 'Select an organization first'} />
				</SelectTrigger>
				<SelectContent>
					<ScrollArea className="h-[200px]">
						{selectedKeys.map((key) => (
							<SelectItem key={key.id} value={key.id.toString()}>
								<div className="flex flex-col">
									<span>{key.name}</span>
									{key.description && <span className="text-xs text-muted-foreground">{key.description}</span>}
									{key.algorithm && (
										<span className="text-xs text-muted-foreground">
											Algorithm: {key.algorithm}
											{key.keySize && ` (${key.keySize} bits)`}
											{key.curve && ` - ${key.curve}`}
										</span>
									)}
								</div>
							</SelectItem>
						))}
					</ScrollArea>
				</SelectContent>
			</Select>
		</div>
	)
}

const DeploymentModal = ({ isOpen, onClose, onDeploy, parameters }: DeploymentModalProps) => {
	// Fetch data for x-source fields
	const { data: organizations } = useQuery({
		...getOrganizationsOptions(),
	})

	const { data: nodes } = useQuery({
		...getNodesOptions(),
	})

	// Dynamically create Zod schema from JSON Schema
	const createZodSchema = (jsonSchema: Record<string, any>) => {
		const schema: Record<string, any> = {}

		Object.entries(jsonSchema.properties || {}).forEach(([key, value]: [string, any]) => {
			// Handle x-source special cases
			if (value['x-source']) {
				switch (value['x-source']) {
					case 'fabric-org':
						schema[key] = z.number()
						break
					case 'key':
						schema[key] = z.number()
						break
					case 'fabric-key':
						schema[key] = z.number()
						break
					case 'fabric-peer':
						if (value.type === 'array') {
							schema[key] = z.array(z.number())
						} else {
							schema[key] = z.number()
						}
						break
					default:
						// Fall back to normal type handling
						break
				}
			} else {
				switch (value.type) {
					case 'string':
						schema[key] = z.string()
						if (value.minLength) schema[key] = schema[key].min(value.minLength)
						if (value.maxLength) schema[key] = schema[key].max(value.maxLength)
						if (value.pattern) schema[key] = schema[key].regex(new RegExp(value.pattern))
						break
					case 'number':
						schema[key] = z.number()
						if (value.minimum) schema[key] = schema[key].min(value.minimum)
						if (value.maximum) schema[key] = schema[key].max(value.maximum)
						break
					case 'boolean':
						schema[key] = z.boolean()
						break
					case 'object':
						if (value.properties) {
							schema[key] = createZodSchema(value)
						} else {
							schema[key] = z.record(z.any())
						}
						break
					case 'array':
						if (value.items) {
							let itemSchema
							if (value.items.type === 'object' && value.items.properties) {
								itemSchema = createZodSchema(value.items)
							} else {
								switch (value.items.type) {
									case 'string':
										itemSchema = z.string()
										break
									case 'number':
										itemSchema = z.number()
										break
									case 'boolean':
										itemSchema = z.boolean()
										break
									default:
										itemSchema = z.any()
								}
							}
							schema[key] = z.array(itemSchema)
							if (value.minItems) schema[key] = schema[key].min(value.minItems)
							if (value.maxItems) schema[key] = schema[key].max(value.maxItems)
						} else {
							schema[key] = z.array(z.any())
						}
						break
				}
			}

			if (!jsonSchema.required?.includes(key)) {
				schema[key] = schema[key].optional()
			}
		})

		return z.object(schema)
	}

	const formSchema = parameters ? createZodSchema(parameters) : z.object({})

	const form = useForm<FormValues>({
		resolver: zodResolver(formSchema),
		defaultValues: {},
	})

	const onSubmit = (values: FormValues) => {
		onDeploy(values)
		onClose()
	}

	// Create form fields dynamically based on JSON Schema
	const renderFormFields = () => {
		if (!parameters?.properties) return null

		return Object.entries(parameters.properties).map(([key, value]: [string, any]) => {
			if (value['x-source']) {
				switch (value['x-source']) {
					case 'fabric-org':
						return (
							<FormField
								key={key}
								control={form.control}
								name={key}
								render={({ field }) => (
									<FormItem>
										<FormLabel className="capitalize">
											{value.title || key}
											{parameters.required?.includes(key) && <span className="text-red-500 ml-1">*</span>}
										</FormLabel>
										<Select value={field.value?.toString()} onValueChange={(val) => field.onChange(Number(val))}>
											<FormControl>
												<SelectTrigger>
													<SelectValue placeholder="Select an organization" />
												</SelectTrigger>
											</FormControl>
											<SelectContent>
												<ScrollArea className="h-[200px]">
													{organizations?.items?.map((org) => (
														<SelectItem key={org.id} value={org.id?.toString()}>
															{org.mspId}
														</SelectItem>
													))}
												</ScrollArea>
											</SelectContent>
										</Select>
										{value.description && <p className="text-sm text-muted-foreground">{value.description}</p>}
										<FormMessage />
									</FormItem>
								)}
							/>
						)
					case 'fabric-key':
						return (
							<FormField
								key={key}
								control={form.control}
								name={key}
								render={({ field }) => (
									<FormItem>
										<FormLabel className="capitalize">
											{value.title || key}
											{parameters.required?.includes(key) && <span className="text-red-500 ml-1">*</span>}
										</FormLabel>
										<FormControl>
											<FabricKeySelect value={field.value as number} onChange={field.onChange} />
										</FormControl>
										{value.description && <p className="text-sm text-muted-foreground">{value.description}</p>}
										<FormMessage />
									</FormItem>
								)}
							/>
						)
					case 'fabric-peer':
						if (value.type === 'array') {
							return (
								<FormField
									key={key}
									control={form.control}
									name={key}
									render={({ field }) => (
										<FormItem className="space-y-4">
											<FormLabel className="capitalize">
												{value.title || key}
												{parameters.required?.includes(key) && <span className="text-red-500 ml-1">*</span>}
											</FormLabel>
											<div className="border rounded-md">
												<ScrollArea className="h-[200px]">
													<div className="p-4 space-y-3">
														{nodes?.items
															?.filter((node) => node.nodeType === 'FABRIC_PEER')
															.map((peer) => (
																<div key={peer.id} className="flex items-center space-x-3 hover:bg-muted/50 rounded-md p-2">
																	<Checkbox
																		id={`peer-${peer.id}`}
																		checked={(field.value as number[])?.includes(peer.id)}
																		onCheckedChange={(checked) => {
																			const currentValue = (field.value as number[]) || []
																			if (checked) {
																				field.onChange([...currentValue, peer.id])
																			} else {
																				field.onChange(currentValue.filter((id) => id !== peer.id))
																			}
																		}}
																		className="h-5 w-5"
																	/>
																	<label
																		htmlFor={`peer-${peer.id}`}
																		className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70 cursor-pointer flex-1"
																	>
																		<div className="font-semibold">{peer.name}</div>
																		<div className="text-muted-foreground text-xs mt-1">ID: {peer.id}</div>
																	</label>
																</div>
															))}
													</div>
												</ScrollArea>
											</div>
											{value.description && <p className="text-sm text-muted-foreground">{value.description}</p>}
											<FormMessage />
										</FormItem>
									)}
								/>
							)
						} else {
							return (
								<FormField
									key={key}
									control={form.control}
									name={key}
									render={({ field }) => (
										<FormItem>
											<FormLabel className="capitalize">
												{value.title || key}
												{parameters.required?.includes(key) && <span className="text-red-500 ml-1">*</span>}
											</FormLabel>
											<Select value={field.value?.toString()} onValueChange={(val) => field.onChange(Number(val))}>
												<FormControl>
													<SelectTrigger>
														<SelectValue placeholder="Select a peer" />
													</SelectTrigger>
												</FormControl>
												<SelectContent>
													<ScrollArea className="h-[200px]">
														{nodes?.items
															?.filter((node) => node.nodeType === 'FABRIC_PEER')
															.map((peer) => (
																<SelectItem key={peer.id} value={peer.id?.toString()}>
																	{peer.name}
																</SelectItem>
															))}
													</ScrollArea>
												</SelectContent>
											</Select>
											{value.description && <p className="text-sm text-muted-foreground">{value.description}</p>}
											<FormMessage />
										</FormItem>
									)}
								/>
							)
						}
					default:
						return null
				}
			}

			// Handle regular form fields
			return (
				<FormField
					key={key}
					control={form.control}
					name={key}
					render={({ field }) => (
						<FormItem>
							<FormLabel className="capitalize">
								{value.title || key}
								{parameters.required?.includes(key) && <span className="text-red-500 ml-1">*</span>}
							</FormLabel>
							<FormControl>
								{value.type === 'boolean' ? (
									<Checkbox checked={field.value as boolean} onCheckedChange={field.onChange} />
								) : (
									<Input
										type={value.type === 'number' ? 'number' : 'text'}
										placeholder={value.description}
										value={field.value as string | number}
										onChange={(e) => {
											const val = value.type === 'number' ? Number(e.target.value) : e.target.value
											field.onChange(val)
										}}
									/>
								)}
							</FormControl>
							{value.description && <p className="text-sm text-muted-foreground">{value.description}</p>}
							<FormMessage />
						</FormItem>
					)}
				/>
			)
		})
	}

	return (
		<Dialog open={isOpen} onOpenChange={onClose}>
			<DialogContent className="sm:max-w-[425px]">
				<DialogHeader>
					<DialogTitle>Deploy Plugin</DialogTitle>
				</DialogHeader>
				<Form {...form}>
					<form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
						{renderFormFields()}
						<DialogFooter>
							<Button type="button" variant="outline" onClick={onClose}>
								Cancel
							</Button>
							<Button type="submit">Deploy</Button>
						</DialogFooter>
					</form>
				</Form>
			</DialogContent>
		</Dialog>
	)
}

export default DeploymentModal
