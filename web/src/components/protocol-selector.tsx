import { FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form'
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group'
import { cn } from '@/lib/utils'
import { Control } from 'react-hook-form'

interface Protocol {
	id: string
	name: string
	logo: string
	category: 'private' | 'public'
	comingSoon?: boolean
}

const protocols: Protocol[] = [
	// Private Networks
	{ id: 'fabric', name: 'Fabric', logo: '/blockchains/fabric.svg', category: 'private' },
	{ id: 'besu', name: 'Besu', logo: '/blockchains/besu_favicon.svg', category: 'private' },
	{ id: 'hiero', name: 'Hiero', logo: '/logos/hiero.svg', category: 'private', comingSoon: true },

	// Public Networks
	{ id: 'avalanche', name: 'Avalanche', logo: '/logos/avalanche.svg', category: 'public', comingSoon: true },
	{ id: 'aptos', name: 'Aptos', logo: '/logos/aptos.svg', category: 'public', comingSoon: true },
	{ id: 'sui', name: 'SUI', logo: '/logos/sui.svg', category: 'public', comingSoon: true },
	{ id: 'ton', name: 'TON', logo: '/logos/ton.svg', category: 'public', comingSoon: true },
]

interface ProtocolSelectorProps {
	control: Control<any>
	name: string
}

export function ProtocolSelector({ control, name }: ProtocolSelectorProps) {
	const privateProtocols = protocols.filter((p) => p.category === 'private')
	const publicProtocols = protocols.filter((p) => p.category === 'public')

	return (
		<FormField
			control={control}
			name={name}
			render={({ field }) => (
				<FormItem className="space-y-3">
					<FormLabel>Select Protocol</FormLabel>
					<FormControl>
						<RadioGroup onValueChange={field.onChange} value={field.value} className="space-y-6">
							{/* Private Networks */}
							<div className="space-y-2">
								<h3 className="text-sm font-medium text-muted-foreground">Private Networks</h3>
								<div className="grid grid-cols-3 gap-4">
									{privateProtocols.map((protocol) => (
										<label
											key={protocol.id}
											className={cn(
												'flex items-center justify-start gap-2 rounded-lg border p-4 cursor-pointer hover:border-primary transition-colors',
												field.value === protocol.id && 'border-primary bg-primary/5',
												protocol.comingSoon && 'opacity-50 cursor-not-allowed'
											)}
										>
											<RadioGroupItem value={protocol.id} id={protocol.id} className="sr-only" disabled={protocol.comingSoon} />
											<img src={protocol.logo} alt={protocol.name} className="h-6 w-6" />
											<span className="text-sm font-medium">
												{protocol.name}
												{protocol.comingSoon && <span className="ml-2 text-xs text-muted-foreground">(Coming Soon)</span>}
											</span>
										</label>
									))}
								</div>
							</div>

							{/* Public Networks */}
							<div className="space-y-2">
								<h3 className="text-sm font-medium text-muted-foreground">Public Networks</h3>
								<div className="grid grid-cols-3 gap-4">
									{publicProtocols.map((protocol) => (
										<label
											key={protocol.id}
											className={cn(
												'flex items-center justify-start gap-2 rounded-lg border p-4 cursor-pointer hover:border-primary transition-colors',
												field.value === protocol.id && 'border-primary bg-primary/5',
												protocol.comingSoon && 'opacity-50 cursor-not-allowed'
											)}
										>
											<RadioGroupItem value={protocol.id} id={protocol.id} className="sr-only" disabled={protocol.comingSoon} />
											<img src={protocol.logo} alt={protocol.name} className="h-6 w-6" />
											<span className="text-sm font-medium">
												{protocol.name}
												{protocol.comingSoon && <span className="ml-2 text-xs text-muted-foreground">(Coming Soon)</span>}
											</span>
										</label>
									))}
								</div>
							</div>
						</RadioGroup>
					</FormControl>
					<FormMessage />
				</FormItem>
			)}
		/>
	)
}
