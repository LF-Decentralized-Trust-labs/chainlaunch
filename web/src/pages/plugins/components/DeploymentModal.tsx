import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog'
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { zodResolver } from '@hookform/resolvers/zod'
import { useForm } from 'react-hook-form'
import { z } from 'zod'

interface DeploymentModalProps {
	isOpen: boolean
	onClose: () => void
	onDeploy: (params: Record<string, unknown>) => void
	parameters?: Record<string, any> // JSON Schema
}

const DeploymentModal = ({ isOpen, onClose, onDeploy, parameters }: DeploymentModalProps) => {
	// Dynamically create Zod schema from JSON Schema
	const createZodSchema = (jsonSchema: Record<string, any>) => {
		const schema: Record<string, any> = {}

		Object.entries(jsonSchema.properties || {}).forEach(([key, value]: [string, any]) => {
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
				// Add more types as needed
			}

			if (!jsonSchema.required?.includes(key)) {
				schema[key] = schema[key].optional()
			}
		})

		return z.object(schema)
	}

	const formSchema = parameters ? createZodSchema(parameters) : z.object({})

	const form = useForm<z.infer<typeof formSchema>>({
		resolver: zodResolver(formSchema),
		defaultValues: {},
	})

	const onSubmit = (values: z.infer<typeof formSchema>) => {
		onDeploy(values)
		onClose()
	}

	// Create form fields dynamically based on JSON Schema
	const renderFormFields = () => {
		if (!parameters?.properties) return null

		return Object.entries(parameters.properties).map(([key, value]: [string, any]) => (
			<FormField
				key={key}
				control={form.control}
				name={key as never}
				render={({ field }) => (
					<FormItem>
						<FormLabel className="capitalize">
							{value.title || key}
							{parameters.required?.includes(key) && <span className="text-red-500 ml-1">*</span>}
						</FormLabel>
						<FormControl>
							{value.type === 'boolean' ? (
								<input type="checkbox" checked={field.value} onChange={(e) => field.onChange(e.target.checked)} />
							) : (
								<Input
									type={value.type === 'number' ? 'number' : 'text'}
									placeholder={value.description}
									{...field}
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
		))
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
