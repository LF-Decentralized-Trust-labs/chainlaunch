import { ModelsProviderResponse } from '@/api/client'
import { OrganizationForm, OrganizationFormValues } from '@/components/forms/organization-form'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog'
import { Plus } from 'lucide-react'

interface CreateOrganizationDialogProps {
	open: boolean
	onOpenChange: (open: boolean) => void
	onSubmit: (data: OrganizationFormValues) => void
	isSubmitting?: boolean
	providers?: ModelsProviderResponse[]
}

export function CreateOrganizationDialog({ open, onOpenChange, onSubmit, isSubmitting, providers }: CreateOrganizationDialogProps) {
	return (
		<Dialog open={open} onOpenChange={onOpenChange}>
			<DialogTrigger asChild>
				<Button>
					<Plus className="h-4 w-4 mr-2" />
					Add Organization
				</Button>
			</DialogTrigger>
			<DialogContent>
				<DialogHeader>
					<DialogTitle>Create New Organization</DialogTitle>
				</DialogHeader>
				<OrganizationForm onSubmit={onSubmit} isSubmitting={isSubmitting} providers={providers} />
			</DialogContent>
		</Dialog>
	)
}
