import { X509Certificate } from '@peculiar/x509'
import { useEffect, useState } from 'react'
import { Buffer } from 'buffer'
interface CertificateThumbprintProps {
	cert: X509Certificate
}

export function CertificateThumbprint({ cert }: CertificateThumbprintProps) {
	const [thumbprint, setThumbprint] = useState<string | null>(null)
	const [subjectName, setSubjectName] = useState<string | null>(null)

	useEffect(() => {
		const loadThumbprint = async () => {
			try {
				const [thumbprintResult, subjectNameResult] = await Promise.all([cert.getThumbprint(), cert.subjectName.getThumbprint()])

				setThumbprint(Buffer.from(thumbprintResult).toString('hex'))
				setSubjectName(Buffer.from(subjectNameResult).toString('hex'))
			} catch (error) {
				console.error('Error loading certificate details:', error)
			}
		}

		loadThumbprint()
	}, [cert])

	if (!thumbprint && !subjectName) return null

	return (
		<>
			{thumbprint && (
				<div>
					<span className="text-muted-foreground">Fingerprint (SHA-1):</span> <code className="text-xs bg-muted px-2 py-1 rounded font-mono">{thumbprint}</code>
				</div>
			)}
			{subjectName && (
				<div>
					<span className="text-muted-foreground">Subject Name Fingerprint:</span> <code className="text-xs bg-muted px-2 py-1 rounded font-mono">{subjectName}</code>
				</div>
			)}
		</>
	)
}
