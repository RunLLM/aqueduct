import {withRequiredAuthInfo} from '@propelauth/react';

function UserInfo(props) {
    console.log('UserInfo props', props);
    const {user} = props;
    return <span>
        <h2>User Info</h2>
        {user && user.pictureUrl && <img src={user.pictureUrl} alt={"profile"} className="pictureUrl" />}
        <pre>user: {JSON.stringify(user, null, 2)}</pre>
    </span>
}

export default withRequiredAuthInfo(UserInfo);