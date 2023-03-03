import React, { useState } from 'react';
import { Box, Dialog, DialogTitle, DialogContent, Typography, TextField, DialogActions, Snackbar, Alert } from '@mui/material';
import { Button } from '../../../primitives/Button.styles';
import { Artifact, ArtifactType, SerializationType } from '../../../../utils/artifacts';
import { useAqueductConsts } from '../../../hooks/useAqueductConsts';
import { handleLoadIntegrations } from '../../../../reducers/integrations';
import { handleGetWorkflow, selectResultIdx } from '../../../../reducers/workflow';
import { useDispatch } from 'react-redux';
import { AppDispatch } from '../../../../stores/store';
import { useNavigate } from 'react-router-dom';
import UserProfile from '../../../../utils/auth';
import { WorkflowDag } from '../../../../utils/workflows';

type RunWorkflowDialogProps = {
    open: boolean;
    setOpen: (boolean) => void;
    user: UserProfile;
    workflowDag: WorkflowDag;
    workflowId: string;
}

const RunWorkflowDialog: React.FC<RunWorkflowDialogProps> = ({ open, setOpen, user, workflowDag, workflowId }) => {
    const { apiAddress } = useAqueductConsts();
    const dispatch: AppDispatch = useDispatch();
    const navigate = useNavigate();
    
    const [showErrorToast, setShowErrorToast] = useState(false);
    const [showSuccessToast, setShowSuccessToast] = useState(false);
  
    const successMessage =
        'Successfully triggered a manual update for this workflow!';
    const [errorMessage, setErrorMessage] = useState(
        'Unable to update this workflow.'
    );

    const name = workflowDag.metadata?.name ?? '';

    const handleSuccessToastClose = async () => {
        setShowSuccessToast(false);

        try {
            await dispatch(handleGetWorkflow({ apiKey: user.apiKey, workflowId }));
            await dispatch(handleLoadIntegrations({ apiKey: user.apiKey }));
            dispatch(selectResultIdx(0));
            navigate(`/workflow/${workflowId}`, { replace: true });
        } catch (error) {
            setErrorMessage(
                `We're having trouble getting the latest workflow. Please try refreshing the page.`
            );
            setShowErrorToast(true);
        }
    };

    const handleErrorToastClose = () => {
        setShowErrorToast(false);
    };


    // This records all the parameters and values that the user wants to overwrite with.
    const [paramNameToValMap, setParamNameToValMap] = useState<{
        [key: string]: string;
    }>({});

    const paramNameToDisplayProps = Object.assign(
        {},
        ...Object.values(workflowDag.operators)
            .filter((operator) => {
                return operator.spec.param !== undefined;
            })
            .map((operator) => {
                // Parameter operators should only have a single output.
                if (operator.outputs.length > 1) {
                    console.error('Parameter operator should not have multiple outputs.');
                }

                // Some types of parameters cannot be easily customized from a textfield on the UI.
                // These types are not json-able and cannot be easily typed as strings.
                const outputArtifact: Artifact =
                    workflowDag.artifacts[operator.outputs[0]];
                const isCustomizable = ![
                    ArtifactType.Table,
                    ArtifactType.Bytes,
                    ArtifactType.Tuple,
                    ArtifactType.Image,
                    ArtifactType.Picklable,
                ].includes(outputArtifact.type);

                let placeholder: string;
                let helperText: string;
                if (isCustomizable) {
                    placeholder = atob(operator.spec.param.val);
                    helperText = '';
                } else {
                    placeholder = '';
                    helperText =
                        outputArtifact.type[0].toUpperCase() +
                        outputArtifact.type.substr(1) +
                        ' type is not yet customizable from the UI.';
                }

                return {
                    [operator.name]: {
                        placeholder: placeholder,
                        isCustomizable: isCustomizable,
                        helperText: helperText,
                    },
                };
            })
    );

    // Returns the map of parameters, from name to spec (which includes the base64-encoded
    // value and serialization_type).
    const serializeParameters = () => {
        const serializedParams = {};
        Object.entries(paramNameToValMap).forEach(([key, strVal]) => {
            // Serialize the user's input string appropriately into base64. The input can either be a
            // 1) number 2) string 3) json.
            try {
                // All jsonable values are serialized as json.
                JSON.parse(strVal);
                serializedParams[key] = {
                    val: btoa(strVal),
                    serialization_type: SerializationType.Json,
                };
            } catch (err) {
                // Non-jsonable values (such as plain strings) are serialized as strings.
                serializedParams[key] = {
                    val: btoa(strVal),
                    serialization_type: SerializationType.String,
                };
            }
        });
        return serializedParams;
    };

    const triggerWorkflowRun = () => {
        const parameters = new FormData();
        parameters.append('parameters', JSON.stringify(serializeParameters()));

        setOpen(false);

        fetch(`${apiAddress}/api/workflow/${workflowDag.workflow_id}/refresh`, {
            method: 'POST',
            headers: {
                'api-key': user.apiKey,
            },
            body: parameters,
        })
            .then((res) => {
                res.json().then((body) => {
                    if (res.ok) {
                        setShowSuccessToast(true);
                    } else {
                        setErrorMessage(`Unable to run this workflow: ${body.error}`);
                        setShowErrorToast(true);
                    }
                });
            })
            .catch(() => {
                setShowErrorToast(true);
            });

        // Reset the overriding parameters map on dialog close.
        setParamNameToValMap({});
    };

    return (
        <>
            <Dialog
                open={open}
                onClose={() => setOpen(false)}
            >
                <DialogTitle>Trigger a Workflow Run?</DialogTitle>
                <DialogContent>
                    <Box sx={{ mb: 2 }}>
                        This will trigger a run of {name} immediately.
                    </Box>

                    {Object.keys(paramNameToDisplayProps).length > 0 && (
                        <Box>
                            <Typography sx={{ mb: 1 }} style={{ fontWeight: 'bold' }}>
                                {' '}
                                Parameters{' '}
                            </Typography>
                            <Typography variant="caption">
                                For json-serializable types like dictionaries or lists, enter the
                                string-serialized representation, without the outer quotes. That
                                is to say, the result of `json.dumps(val)`.
                            </Typography>
                        </Box>
                    )}
                    {Object.keys(paramNameToDisplayProps).map((paramName) => {
                        return (
                            <Box key={paramName}>
                                <Typography>
                                    <small>{paramName}</small>
                                </Typography>
                                <TextField
                                    fullWidth
                                    disabled={!paramNameToDisplayProps[paramName].isCustomizable}
                                    helperText={paramNameToDisplayProps[paramName].helperText}
                                    placeholder={paramNameToDisplayProps[paramName].placeholder}
                                    onChange={(e) => {
                                        paramNameToValMap[paramName] = e.target.value;
                                        setParamNameToValMap(paramNameToValMap);
                                    }}
                                    size="small"
                                />
                            </Box>
                        );
                    })}
                </DialogContent>
                <DialogActions>
                    <Button
                        color="secondary"
                        onClick={() => setOpen(false)}
                    >
                        Cancel
                    </Button>
                    <Button color="primary" onClick={() => triggerWorkflowRun()}>
                        Run
                    </Button>
                </DialogActions>
            </Dialog >
            <Snackbar
                anchorOrigin={{ vertical: 'top', horizontal: 'center' }}
                open={showSuccessToast}
                onClose={handleSuccessToastClose}
                key={'workflowheader-success-snackbar'}
                autoHideDuration={4000}
            >
                <Alert
                    onClose={handleSuccessToastClose}
                    severity="success"
                    sx={{ width: '100%' }}
                >
                    {successMessage}
                </Alert>
            </Snackbar>
            <Snackbar
                anchorOrigin={{ vertical: 'top', horizontal: 'center' }}
                open={showErrorToast}
                onClose={handleErrorToastClose}
                key={'workflowheader-error-snackbar'}
                autoHideDuration={4000}
            >
                <Alert
                    onClose={handleErrorToastClose}
                    severity="error"
                    sx={{ width: '100%' }}
                >
                    {errorMessage}
                </Alert>
            </Snackbar>
        </>
    );
};

export { RunWorkflowDialog };