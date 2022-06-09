export interface UserProfile {
  name?: string;
  email?: string;
  sub?: string;
  apiKey?: string;
  updated_at?: string;
  email_verified?: boolean;
  nickname?: string;
  picture?: string;
  organizationId?: string;
  userRole?: string;
  given_name?: string;
  family_name?: string;
}

export default UserProfile;
