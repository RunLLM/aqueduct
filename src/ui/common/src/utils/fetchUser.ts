import { apiAddress } from '../components/hooks/useAqueductConsts';
import UserProfile from './auth';

export default async function fetchUser(
  apiKey: string
): Promise<{ success: boolean; user?: UserProfile }> {
  try {
    const response = await fetch(`${apiAddress}/api/user`, {
      method: 'GET',
      headers: {
        'api-key': apiKey,
      },
    });

    if (!response.ok) {
      return { success: false, user: undefined };
    }

    const body = await response.json();

    return {
      success: true,
      user: {
        apiKey: apiKey,
        email: body.email,
        email_verified: true,
        name: 'aqueduct user',
        nickname: 'aqueduct user',
        organizationId: body.organization_id,
        picture: undefined,
        sub: 'default',
        updated_at: 'default',
        userRole: body.role,
      },
    };
  } catch (error) {
    return { success: false, user: undefined };
  }
}
