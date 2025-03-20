import { ModelsProviderResponse } from '@/api/client'
import { Button } from '@/components/ui/button'
import { Command, CommandEmpty, CommandGroup, CommandInput, CommandItem, CommandList } from '@/components/ui/command'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { Skeleton } from '@/components/ui/skeleton'

interface ProviderFilterProps {
	providers?: ModelsProviderResponse[]
	selectedProvider: number | null
	onProviderChange: (providerId: number | null) => void
	isLoading?: boolean
}

export function ProviderFilter({ providers, selectedProvider, onProviderChange, isLoading }: ProviderFilterProps) {
	return (
		<Popover>
			<PopoverTrigger asChild>
				<Button variant="outline" role="combobox" className="w-[200px] justify-between" disabled={isLoading}>
					{isLoading ? <Skeleton className="h-4 w-[160px]" /> : <>{selectedProvider ? providers?.find((p) => p.id === selectedProvider)?.name || 'Select Provider' : 'All Providers'}</>}
				</Button>
			</PopoverTrigger>
			<PopoverContent className="w-[200px] p-0">
				<Command>
					<CommandInput placeholder="Search providers..." />
					<CommandList>
						<CommandEmpty>No providers found.</CommandEmpty>
						<CommandGroup>
							<CommandItem value="" onSelect={() => onProviderChange(null)}>
								All Providers
							</CommandItem>
							{providers?.map((provider) => (
								<CommandItem key={provider.id} value={provider.id?.toString() || ''} onSelect={() => onProviderChange(provider.id || null)}>
									{provider.name}
								</CommandItem>
							))}
						</CommandGroup>
					</CommandList>
				</Command>
			</PopoverContent>
		</Popover>
	)
}
