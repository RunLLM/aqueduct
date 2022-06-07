//import { UserProfile } from '@utils/auth';
//import fetchUser from
import { useEffect, useState } from 'react';
import cookie from 'react-cookies';

import UserProfile from '../../utils/auth';
import fetchUser from '../../utils/fetchUser';

export default function useUser(): {
  success: boolean;
  loading: boolean;
  user?: UserProfile;
} {
  const apiKey = cookie.load('aqueduct-api-key');
  const [user, setUser] = useState<UserProfile>(undefined);
  const [success, setSuccess] = useState(false);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const getUser = async () => {
      try {
        setLoading(true);
        const { success, user } = await fetchUser(apiKey);
        setSuccess(success);
        setUser(user);
        setLoading(false);
      } catch (error) {
        setSuccess(false);
        setUser(undefined);
        setLoading(false);
      }
    };

    getUser();
  }, []);

  return { success, loading, user };
}
