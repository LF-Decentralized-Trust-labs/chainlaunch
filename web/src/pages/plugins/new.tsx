import { zodResolver } from '@hookform/resolvers/zod'
import { useForm } from 'react-hook-form'
import { useNavigate } from 'react-router-dom'
import * as z from 'zod'
import { parse, stringify } from 'yaml'

import { Button } from '@/components/ui/button'
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { postPlugins } from '@/api/client/sdk.gen'
import { toast } from 'sonner'
import { Textarea } from '@/components/ui/textarea'

const defaultYaml = `apiVersion: dev.chainlaunch/v1
kind: Plugin
metadata:
  name: my-plugin
  version: "1.0"
spec:
  dockerCompose:
    contents: |
      version: '2.2'
      services:
        app:
          image: nginx:latest
  parameters:
    $schema: http://json-schema.org/draft-07/schema#
    type: object
    properties: {}
    required: []`

const formSchema = z.object({
	yaml: z.string()
		.min(1, "YAML is required")
		.refine((yaml) => {
			try {
				const parsed = parse(yaml)
				// Add basic validation for required fields
				if (!parsed.apiVersion || !parsed.kind || !parsed.metadata?.name) {
					return false
				}
				return true
			} catch (e) {
				return false
			}
		}, "Invalid YAML format or missing required fields")
})

type FormValues = z.infer<typeof formSchema>

export default function NewPluginPage() {
	const navigate = useNavigate()

	const form = useForm<FormValues>({
		resolver: zodResolver(formSchema),
		defaultValues: {
			yaml: defaultYaml
		},
	})

	async function onSubmit(data: FormValues) {
		try {
			const pluginData = parse(data.yaml)
			const response = await postPlugins({
				body: pluginData,
			})

			toast.success('Plugin created successfully')
			navigate(`/plugins/${pluginData.metadata.name}`)
		} catch (error) {
			toast.error('Failed to create plugin. Please check your YAML format and try again.')
		}
	}

	return (
		<div className="container mx-auto py-6">
			<Card>
				<CardHeader>
					<CardTitle>Create New Plugin</CardTitle>
					<CardDescription>
						Create a new plugin by providing the plugin configuration in YAML format
					</CardDescription>
				</CardHeader>
				<CardContent>
					<Form {...form}>
						<form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
							<FormField
								control={form.control}
								name="yaml"
								render={({ field }) => (
									<FormItem>
										<FormLabel>Plugin Configuration (YAML)</FormLabel>
										<FormControl>
											<Textarea
												{...field}
												className="font-mono min-h-[500px]"
												placeholder="Enter your plugin YAML configuration..."
											/>
										</FormControl>
										<FormMessage />
									</FormItem>
								)}
							/>

							<div className="flex justify-end space-x-4">
								<Button 
									type="button" 
									variant="outline" 
									onClick={() => navigate('/plugins')}
								>
									Cancel
								</Button>
								<Button type="submit">Create Plugin</Button>
							</div>
						</form>
					</Form>
				</CardContent>
			</Card>
		</div>
	)
}
