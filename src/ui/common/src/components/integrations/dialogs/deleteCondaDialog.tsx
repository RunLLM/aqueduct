// import React, {useEffect, useState} from "react";
// import UserProfile from "../../../utils/auth";
// import {IntegrationConfig, Service} from "../../../utils/integrations";
// import Dialog from "@mui/material/Dialog";
// import {isFailed, isLoading, isSucceeded} from "../../../utils/shared";
// import {Alert, DialogActions, DialogContent} from "@mui/material";
// import Button from "@mui/material/Button";
// import {LoadingButton} from "@mui/lab";
// import {handleDeleteIntegration, resetDeletionStatus} from "../../../reducers/integration";
// import {AppDispatch, RootState} from "../../../stores/store";
// import {useDispatch, useSelector} from "react-redux";
// import {useNavigate} from "react-router-dom";
//
// type Props = {
//   user: UserProfile;
//   integrationId: string;
//   integrationName: string;
//   integrationType: Service;
//   config: IntegrationConfig;
//   onCloseDialog: () => void;
// };
//
//
// // This is a much simpler version of the deletion dialog, meant to avoid all the intricacies and guardrails
// // for the other resources, since can be deleted whenever.
// const DeleteCondaDialog: React.FC<Props> = ({
//     user,
//     integrationId,
//     integrationName,
//     integrationType,
//     config,
//     onCloseDialog,
// }) => {
//   const dispatch: AppDispatch = useDispatch();
//   const navigate = useNavigate();
//   const [isConnecting, setIsConnecting] = useState(false);
//
//   const deleteIntegrationStatus = useSelector(
//       (state: RootState) => state.integrationReducer.deletionStatus
//   );
//
//
//   useEffect(() => {
//     if (!isLoading(deleteIntegrationStatus)) {
//       setIsConnecting(false);
//     }
//
//     if (isSucceeded(deleteIntegrationStatus)) {
//       navigate('/resources', {
//         state: {
//           deleteIntegrationStatus: deleteIntegrationStatus,
//           deleteIntegrationName: integrationName,
//         },
//       });
//     }
//   }, [deleteIntegrationStatus, integrationName, navigate]);
//
//   const confirmConnect = () => {
//     setIsConnecting(true);
//     dispatch(
//         handleDeleteIntegration({
//           apiKey: user.apiKey,
//           integrationId: integrationId,
//         })
//     );
//   };
//
//   return (
//       <>
//         <Dialog
//             open={!deleteIntegrationStatus || !isFailed(deleteIntegrationStatus)}
//             onClose={onCloseDialog}
//             maxWidth="lg"
//         >
//           <DialogContent>
//             Are you sure you want to delete the resource?
//           </DialogContent>
//           <DialogActions>
//             <Button onClick={onCloseDialog}>Cancel</Button>
//             <LoadingButton
//                 autoFocus
//                 onClick={confirmConnect}
//                 loading={isConnecting}
//             >
//               Confirm
//             </LoadingButton>
//           </DialogActions>
//         </Dialog>
//         <Dialog
//             open={isFailed(deleteIntegrationStatus)}
//             onClose={onCloseDialog}
//             maxWidth="lg"
//         >
//           {deleteIntegrationStatus && isFailed(deleteIntegrationStatus) && (
//               <Alert severity="error" sx={{ margin: 2 }}>
//                 Integration deletion failed with error:
//                 <br></br>
//                 <pre>{deleteIntegrationStatus.err}</pre>
//               </Alert>
//           )}
//           <DialogActions>
//             <Button
//                 onClick={() => {
//                   onCloseDialog();
//                   dispatch(resetDeletionStatus());
//                 }}
//             >
//               Dismiss
//             </Button>
//           </DialogActions>
//         </Dialog>
//       </>
//   );
//
// }
