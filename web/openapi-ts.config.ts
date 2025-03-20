export default {
	client: '@hey-api/client-fetch',
	input: 'http://localhost:8100/swagger/doc.json',
	output: 'src/api/client',
	plugins: ['@tanstack/react-query'],
}
