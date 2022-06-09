import { useEffect, useState } from 'react';
import { useCookies } from 'react-cookie';

import UserProfile from '../../utils/auth';
import fetchUser from '../../utils/fetchUser';

export default function useUser(): {
  success: boolean;
  loading: boolean;
  user?: UserProfile;
} {
  const [cookies, setCookie, removeCookie] = useCookies(['aqueduct-api-key']);
  const apiKey = cookies['aqueduct-api-key'];
  const [user, setUser] = useState<UserProfile>(undefined);
  const [success, setSuccess] = useState(false);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const getUser = async () => {
      try {
        setLoading(true);
        const { success, user } = await fetchUser(apiKey);

        setUser(user);
        setSuccess(success);
        setCookie('aqueduct-api-key', apiKey, { path: '/' });
        setLoading(false);
      } catch (error) {
        setSuccess(false);
        setUser(undefined);
        setCookie('aqueduct-api-key', apiKey, { path: '/' });
        setLoading(false);
      }
    };

    getUser();
  }, []);

  return { success, loading, user };
}
