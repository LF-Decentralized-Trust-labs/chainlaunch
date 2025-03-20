import { Button } from '@/components/ui/button'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from '@/components/ui/dropdown-menu'
import { BesuIcon } from '@/components/icons/besu-icon'
import { FabricIcon } from '@/components/icons/fabric-icon'
import { Plus, ChevronDown } from 'lucide-react'
import { useNavigate } from 'react-router-dom'

export function CreateNodeButton() {
	const navigate = useNavigate()

	return (
		<DropdownMenu>
			<DropdownMenuTrigger asChild>
				<Button>
					<Plus className="mr-2 h-4 w-4" />
					Create Node
					<ChevronDown className="ml-2 h-4 w-4" />
				</Button>
			</DropdownMenuTrigger>
			<DropdownMenuContent align="end">
				<DropdownMenuItem onClick={() => navigate('/nodes/fabric/create')}>
					<FabricIcon className="mr-2 h-4 w-4" />
					Fabric Node
				</DropdownMenuItem>
				<DropdownMenuItem onClick={() => navigate('/nodes/besu/create')}>
					<BesuIcon className="mr-2 h-4 w-4" />
					Besu Node
				</DropdownMenuItem>
			</DropdownMenuContent>
		</DropdownMenu>
	)
}
