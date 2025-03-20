import { z } from 'zod'

export const targetFormSchema = z.object({
	name: z.string().min(1),
	endpoint: z.string().min(1),
	type: z.literal('S3'),
	accessKeyId: z.string().min(1),
	secretKey: z.string().min(1),
	bucketName: z.string().min(1),
	bucketPath: z.string().min(1),
	region: z.string().min(1),
	forcePathStyle: z.boolean().optional(),
})

export type TargetFormValues = z.infer<typeof targetFormSchema> 