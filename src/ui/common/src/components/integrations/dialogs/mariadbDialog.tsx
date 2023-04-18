// import { yupResolver } from '@hookform/resolvers/yup';
import Box from '@mui/material/Box';
import TextField from '@mui/material/Textfield';
import Typography from '@mui/material/Typography';
import React from 'react';
import { useFormContext } from 'react-hook-form';
// import * as Yup from 'yup';

import { MariaDbConfig } from '../../../utils/integrations';
import { Button } from '../../primitives/Button.styles';

const Placeholders: MariaDbConfig = {
  host: '127.0.0.1',
  port: '3306',
  database: 'aqueduct-db',
  username: 'aqueduct',
  password: '********',
};

type Props = {
  onUpdateField: (field: keyof MariaDbConfig, value: string) => void;
  value?: MariaDbConfig;
  editMode: boolean;
};

export const MariaDbDialog: React.FC<Props> = ({
  onUpdateField,
  value,
  editMode,
}) => {
  //const validationSchema = Yup.object().shape({
    //host: Yup.string().required('Please enter a host url.'),
    //port: Yup.string().required('Please enter a port number.'),
    //database: Yup.string().required('Please enter a database name.'),
    //username: Yup.string().required('Please enter a username.'),
    //password: Yup.string().required('Please enter a password.'),
  //});

  const onSubmit = (data) => {
    console.log(JSON.stringify(data, null, 2));
  };

  // const {
  //   register,
  //   control,
  //   handleSubmit,
  //   formState: { errors },
  // } = useForm({
  //   resolver: yupResolver(validationSchema),
  // });

  const { register, errors, formState } = useFormContext();

  console.log('formState from context: ', formState);
  console.log('errors from context: ', formState.errors)
  console.log('touchedFields: ', formState.touchedFields)

  return (
    <Box sx={{ mt: 2 }}>
      {/* <IntegrationTextInputField
        label={'Host*'}
        description={'The hostname or IP address of the MariaDB server.'}
        spellCheck={false}
        required={true}
        placeholder={Placeholders.host}
        onChange={(event) => onUpdateField('host', event.target.value)}
        value={value?.host ?? ''}
        disabled={editMode}
        warning={editMode ? undefined : readOnlyFieldWarning}
        disableReason={editMode ? readOnlyFieldDisableReason : undefined}
      /> */}

      <TextField
        required
        id="host"
        name="host"
        label="Host"
        fullWidth
        margin="dense"
        error={errors?.host ? true : false}
        {...register('host')}
      />
      <Typography variant="inherit" color="textSecondary">
        {errors?.host?.message}
      </Typography>

      {/* <IntegrationTextInputField
        label={'Port*'}
        description={'The port number of the MariaDB server.'}
        spellCheck={false}
        required={true}
        placeholder={Placeholders.port}
        onChange={(event) => onUpdateField('port', event.target.value)}
        value={value?.port ?? ''}
        disabled={editMode}
        warning={editMode ? undefined : readOnlyFieldWarning}
        disableReason={editMode ? readOnlyFieldDisableReason : undefined}
      /> */}

      <TextField
        required
        id="port"
        name="port"
        label="Port"
        fullWidth
        margin="dense"
        error={errors?.port ? true : false}
        {...register('port')}
      />
      <Typography variant="inherit" color="textSecondary">
        {errors?.port?.message}
      </Typography>

      {/* <IntegrationTextInputField
        label={'Database*'}
        description={'The name of the specific database to connect to.'}
        spellCheck={false}
        required={true}
        placeholder={Placeholders.database}
        onChange={(event) => onUpdateField('database', event.target.value)}
        value={value?.database ?? ''}
        disabled={editMode}
        warning={editMode ? undefined : readOnlyFieldWarning}
        disableReason={editMode ? readOnlyFieldDisableReason : undefined}
      /> */}

      {/* <IntegrationTextInputField
        spellCheck={false}
        required={true}
        label="Username*"
        description="The username of a user with access to the above database."
        placeholder={Placeholders.username}
        onChange={(event) => onUpdateField('username', event.target.value)}
        value={value?.username ?? ''}
      /> */}

      {/* <IntegrationTextInputField
        spellCheck={false}
        required={true}
        label="Password*"
        description="The password corresponding to the above username."
        placeholder={Placeholders.password}
        type="password"
        onChange={(event) => onUpdateField('password', event.target.value)}
        value={value?.password ?? ''}
      /> */}

      <Button
        variant="contained"
        color="primary"
        onClick={() => console.log('register clicked')}
      >
        Register
      </Button>
    </Box>
  );
};

export const isMariaDBConfigComplete = (config: MariaDbConfig): boolean => {
  return (
    !!config.database &&
    !!config.host &&
    !!config.password &&
    !!config.port &&
    !!config.username &&
    !!config.port
  );
};
