import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Copy } from 'lucide-react'
import { useState } from 'react'
import { toast } from 'sonner'
import { stringify } from 'yaml'

interface YamlViewerProps {
	yaml: any
	dialogOpen: boolean
	setDialogOpen: (open: boolean) => void
}

export function YamlViewer({ yaml, dialogOpen, setDialogOpen }: YamlViewerProps) {

	const copyToClipboard = () => {
		const yamlString = stringify(yaml)
		navigator.clipboard.writeText(yamlString)
		toast.success('YAML copied to clipboard')
	}

	return (
		<>
			<Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
				<DialogContent className="max-w-3xl">
					<DialogHeader>
						<DialogTitle>Plugin YAML Specification</DialogTitle>
					</DialogHeader>
					<div className="relative">
						<Button variant="ghost" size="icon" className="absolute right-2 top-2" onClick={copyToClipboard}>
							<Copy className="h-4 w-4" />
						</Button>
						<pre className="text-sm font-mono bg-muted p-4 rounded-lg overflow-x-auto whitespace-pre-wrap break-all max-h-[70vh] overflow-y-auto">{stringify(yaml)}</pre>
					</div>
				</DialogContent>
			</Dialog>
		</>
	)
}
