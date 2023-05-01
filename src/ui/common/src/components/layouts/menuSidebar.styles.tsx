export const menuSidebar = {
  maxWidth: '80px',
  width: '80px',
  height: '100%',
  maxHeight: '100vh',
  paddingBottom: '8px',
  display: 'flex',
  flexDirection: 'column',
  backgroundColor: '#002f5e' /* theme.colors.blue900 */,
  position: 'fixed',
  zIndex: 1,
};

export const menuSidebarContent = {
  display: 'flex',
  flexDirection: 'column',
  flexGrow: 1,
  width: '100%',
  paddingLeft: '8px',
  paddingRight: '8px',
  marginTop: '8px',
};

export const menuSidebarFooter = {
  display: 'flex',
  flexDirection: 'column',
  width: '100%',
  paddingLeft: '8px',
  paddingRight: '8px',
};

export const menuSidebarLinksWrapper = {
  flexGrow: 1,
  height: '100%',
};

export const menuSidebarLink = {
  /* this is fixed to 225 - 13 - 13, the sidebar width with padding-X removed */
  minWidth: '64px',
  width: '64px',
  maxWidth: '64px',
  marginTop: '8px',
  marginBottom: '8px',
};

export const menuSidebarLogoLink = {
  /* contains logo of size 48px */
  width: '100%',
  height: '64px',
  /* ensures border width is 80px, which maps sidebar */
  paddingLeft: '16px',
  paddingRight: '16px',

  /* ensures height is 64px, which maps navbar */
  paddingTop: '8px',
  paddingBottom: '6px',
  borderBottom: '2px solid #E6E8F0' /* gray.300 */,
};

export const menuSidebarIcon = {
  /* All icons in sidebar button are fixed to 24. */
  minWidth: '24px',
  minHeight: '24px',
  maxWidth: '24px',
  maxHeight: '24px',
};

export const navbarIcon = {
  /* All icons in sidebar button are fixed to 24. */
  minWidth: '24px',
  minHeight: '24px',
  maxWidth: '24px',
  maxHeight: '24px',
  color: '#002f5e' /* blue900 */,
};

export const userAvatarImg = {
  width: '24px',
  height: '24px',
  maxWidth: '24px',
  maxHeight: '24px',
  borderRadius: '99999px',
};

export const notificationAlert = {
  display: 'flex',
  backgroundColor: '#d14343' /* red500 */,
  /* All icons in sidebar button are fixed to 24 */
  minWidth: '24px',
  width: '24px',
  maxWidth: '24px',
  height: '24px',
  borderRadius: '8px',
  justifyContent: 'center',
  alignItems: 'center',
};
