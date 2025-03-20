import { ServiceFabricOrdererProperties } from '@/api/client'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Separator } from '@/components/ui/separator'

interface FabricOrdererConfigProps {
	config: ServiceFabricOrdererProperties
}

export function FabricOrdererConfig({ config }: FabricOrdererConfigProps) {
	return (
		<Card>
			<CardHeader>
				<CardTitle>Fabric Orderer Configuration</CardTitle>
				<CardDescription>Orderer-specific node settings</CardDescription>
			</CardHeader>
			<CardContent className="space-y-6">
				<div className="grid grid-cols-2 gap-4">
					<div>
						<p className="text-sm font-medium text-muted-foreground">Organization</p>
						<p>MSP ID: {config.mspId}</p>
						<p className="text-sm text-muted-foreground">Org ID: {config.organizationId}</p>
					</div>
					<div>
						<p className="text-sm font-medium text-muted-foreground">Key IDs</p>
						<p>Sign Key: {config.signKeyId}</p>
						<p>TLS Key: {config.tlsKeyId}</p>
					</div>
				</div>

				<Separator />

				<div className="space-y-2">
					<p className="text-sm font-medium text-muted-foreground">Network Configuration</p>
					<div className="grid grid-cols-2 gap-4">
						<div>
							<p className="text-sm font-medium text-muted-foreground">Listen Address</p>
							<p className="text-sm">{config.listenAddress}</p>
						</div>
						<div>
							<p className="text-sm font-medium text-muted-foreground">Admin Address</p>
							<p className="text-sm">{config.adminAddress}</p>
						</div>
						<div>
							<p className="text-sm font-medium text-muted-foreground">Operations Address</p>
							<p className="text-sm">{config.operationsAddress}</p>
						</div>
					</div>
				</div>

				{config.externalEndpoint && (
					<>
						<Separator />
						<div>
							<p className="text-sm font-medium text-muted-foreground">External Endpoint</p>
							<p className="text-sm">{config.externalEndpoint}</p>
						</div>
					</>
				)}

				{config.domainNames && config.domainNames.length > 0 && (
					<>
						<Separator />
						<div className="space-y-2">
							<p className="text-sm font-medium text-muted-foreground">Domains</p>
							<div className="flex flex-wrap gap-2">
								{config.domainNames.map((domain) => (
									<Badge key={domain} variant="outline">
										{domain}
									</Badge>
								))}
							</div>
						</div>
					</>
				)}
			</CardContent>
		</Card>
	)
}
