import { RedocStandalone } from 'redoc'
import { useEffect, useState } from 'react'

const ApiDocumentationPage = () => {
	const [spec, setSpec] = useState<object | null>(null)
	const [loading, setLoading] = useState(true)
	const [error, setError] = useState<string | null>(null)

	useEffect(() => {
		const fetchSwaggerDoc = async () => {
			try {
				setLoading(true)
				const response = await fetch('/api/swagger/doc.json')
				if (!response.ok) {
					throw new Error(`Failed to fetch API documentation: ${response.statusText}`)
				}
				const data = await response.json()
				setSpec(data)
			} catch (err) {
				console.error('Error fetching API documentation:', err)
				setError(err instanceof Error ? err.message : 'Failed to load API documentation')
			} finally {
				setLoading(false)
			}
		}

		fetchSwaggerDoc()
	}, [])

	if (loading) {
		return <div className="container p-8 text-center">Loading API documentation...</div>
	}

	if (error) {
		return <div className="container p-8 text-center text-red-500">Error: {error}</div>
	}

	if (!spec) {
		return <div className="container p-8 text-center">No API documentation available</div>
	}
	return (
		<RedocStandalone
			spec={spec}
			options={{
				theme: {
					colors: {
						primary: {
							main: '#2196f3',
						},
					},
					typography: {
						fontSize: '16px',
						fontFamily: 'Inter, -apple-system, BlinkMacSystemFont, sans-serif',
					},
					sidebar: {
						backgroundColor: '#f5f5f5',
					},
				},
				nativeScrollbars: true,
				hideDownloadButton: false,
				expandResponses: '200,201',
				jsonSampleExpandLevel: 3,
			}}
		/>
	)
}

export default ApiDocumentationPage
