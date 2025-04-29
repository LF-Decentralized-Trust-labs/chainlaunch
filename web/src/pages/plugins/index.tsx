import { deletePluginsByNameMutation, getPluginsOptions } from '@/api/client/@tanstack/react-query.gen'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import {
	AlertDialog,
	AlertDialogAction,
	AlertDialogCancel,
	AlertDialogContent,
	AlertDialogDescription,
	AlertDialogFooter,
	AlertDialogHeader,
	AlertDialogTitle,
} from "@/components/ui/alert-dialog"
import { useMutation, useQuery } from '@tanstack/react-query'
import { Trash2 } from 'lucide-react'
import { Link } from 'react-router-dom'
import { toast } from 'sonner'
import { useState } from 'react'

const PluginsPage = () => {
	const [pluginToDelete, setPluginToDelete] = useState<string | null>(null)
	
	// Fetch plugins
	const { data: plugins, isLoading, error, refetch } = useQuery(getPluginsOptions())

	// Delete mutation
	const deleteMutation = useMutation({
		...deletePluginsByNameMutation(),
		onSuccess: () => {
			toast.success('Plugin deleted successfully')
			refetch()
		},
		onError: (error) => {
			toast.error(`Failed to delete plugin: ${error}`)
		},
	})

	const handleDelete = (name: string) => {
		deleteMutation.mutate({ path: { name } })
		setPluginToDelete(null)
	}

	if (isLoading) return <div className="container p-8">Loading...</div>
	if (error) return <div className="container p-8 text-red-500">Error loading plugins</div>

	return (
		<div className="container p-8">
			<div className="flex justify-between items-center mb-6">
				<h1 className="text-2xl font-bold">Plugins</h1>
				<Button asChild>
					<Link to="/plugins/new">Create Plugin</Link>
				</Button>
			</div>

			<div className="grid gap-4">
				{plugins?.map((plugin) => (
					<Card key={plugin.metadata?.name}>
						<CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
							<CardTitle className="text-lg">
								<Link 
									to={`/plugins/${plugin.metadata?.name}`} 
									className="hover:text-primary transition-colors"
								>
									{plugin.metadata?.name}
								</Link>
							</CardTitle>
							<Button
								variant="destructive"
								size="icon"
								onClick={() => setPluginToDelete(plugin.metadata?.name || null)}
							>
								<Trash2 className="h-4 w-4" />
							</Button>
						</CardHeader>
						<CardContent>
							<p className="text-muted-foreground">{plugin.metadata?.description}</p>
						</CardContent>
					</Card>
				))}
			</div>

			<AlertDialog 
				open={!!pluginToDelete} 
				onOpenChange={(open) => !open && setPluginToDelete(null)}
			>
				<AlertDialogContent>
					<AlertDialogHeader>
						<AlertDialogTitle>Are you sure?</AlertDialogTitle>
						<AlertDialogDescription>
							This will permanently delete the plugin "{pluginToDelete}".
							This action cannot be undone.
						</AlertDialogDescription>
					</AlertDialogHeader>
					<AlertDialogFooter>
						<AlertDialogCancel>Cancel</AlertDialogCancel>
						<AlertDialogAction
							className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
							onClick={() => pluginToDelete && handleDelete(pluginToDelete)}
						>
							Delete
						</AlertDialogAction>
					</AlertDialogFooter>
				</AlertDialogContent>
			</AlertDialog>
		</div>
	)
}

export default PluginsPage
