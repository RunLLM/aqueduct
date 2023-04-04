import { emptySplitApi as api } from "./placeholder";
const injectedRtkApi = api.injectEndpoints({
  endpoints: (build) => ({
    uploadFile: build.mutation<UploadFileApiResponse, UploadFileApiArg>({
      query: (queryArg) => ({
        url: `/pet/${queryArg.petId}/uploadImage`,
        method: "POST",
        body: queryArg.body,
      }),
    }),
    addPet: build.mutation<AddPetApiResponse, AddPetApiArg>({
      query: (queryArg) => ({
        url: `/pet`,
        method: "POST",
        body: queryArg.pet,
      }),
    }),
    updatePet: build.mutation<UpdatePetApiResponse, UpdatePetApiArg>({
      query: (queryArg) => ({ url: `/pet`, method: "PUT", body: queryArg.pet }),
    }),
    findPetsByStatus: build.query<
      FindPetsByStatusApiResponse,
      FindPetsByStatusApiArg
    >({
      query: (queryArg) => ({
        url: `/pet/findByStatus`,
        params: { status: queryArg.status },
      }),
    }),
    findPetsByTags: build.query<
      FindPetsByTagsApiResponse,
      FindPetsByTagsApiArg
    >({
      query: (queryArg) => ({
        url: `/pet/findByTags`,
        params: { tags: queryArg.tags },
      }),
    }),
    getPetById: build.query<GetPetByIdApiResponse, GetPetByIdApiArg>({
      query: (queryArg) => ({ url: `/pet/${queryArg.petId}` }),
    }),
    updatePetWithForm: build.mutation<
      UpdatePetWithFormApiResponse,
      UpdatePetWithFormApiArg
    >({
      query: (queryArg) => ({
        url: `/pet/${queryArg.petId}`,
        method: "POST",
        body: queryArg.body,
      }),
    }),
    deletePet: build.mutation<DeletePetApiResponse, DeletePetApiArg>({
      query: (queryArg) => ({
        url: `/pet/${queryArg.petId}`,
        method: "DELETE",
        headers: { api_key: queryArg.apiKey },
      }),
    }),
    placeOrder: build.mutation<PlaceOrderApiResponse, PlaceOrderApiArg>({
      query: (queryArg) => ({
        url: `/store/order`,
        method: "POST",
        body: queryArg.order,
      }),
    }),
    getOrderById: build.query<GetOrderByIdApiResponse, GetOrderByIdApiArg>({
      query: (queryArg) => ({ url: `/store/order/${queryArg.orderId}` }),
    }),
    deleteOrder: build.mutation<DeleteOrderApiResponse, DeleteOrderApiArg>({
      query: (queryArg) => ({
        url: `/store/order/${queryArg.orderId}`,
        method: "DELETE",
      }),
    }),
    getInventory: build.query<GetInventoryApiResponse, GetInventoryApiArg>({
      query: () => ({ url: `/store/inventory` }),
    }),
    createUsersWithArrayInput: build.mutation<
      CreateUsersWithArrayInputApiResponse,
      CreateUsersWithArrayInputApiArg
    >({
      query: (queryArg) => ({
        url: `/user/createWithArray`,
        method: "POST",
        body: queryArg.body,
      }),
    }),
    createUsersWithListInput: build.mutation<
      CreateUsersWithListInputApiResponse,
      CreateUsersWithListInputApiArg
    >({
      query: (queryArg) => ({
        url: `/user/createWithList`,
        method: "POST",
        body: queryArg.body,
      }),
    }),
    getUserByName: build.query<GetUserByNameApiResponse, GetUserByNameApiArg>({
      query: (queryArg) => ({ url: `/user/${queryArg.username}` }),
    }),
    updateUser: build.mutation<UpdateUserApiResponse, UpdateUserApiArg>({
      query: (queryArg) => ({
        url: `/user/${queryArg.username}`,
        method: "PUT",
        body: queryArg.user,
      }),
    }),
    deleteUser: build.mutation<DeleteUserApiResponse, DeleteUserApiArg>({
      query: (queryArg) => ({
        url: `/user/${queryArg.username}`,
        method: "DELETE",
      }),
    }),
    loginUser: build.query<LoginUserApiResponse, LoginUserApiArg>({
      query: (queryArg) => ({
        url: `/user/login`,
        params: { username: queryArg.username, password: queryArg.password },
      }),
    }),
    logoutUser: build.query<LogoutUserApiResponse, LogoutUserApiArg>({
      query: () => ({ url: `/user/logout` }),
    }),
    createUser: build.mutation<CreateUserApiResponse, CreateUserApiArg>({
      query: (queryArg) => ({
        url: `/user`,
        method: "POST",
        body: queryArg.user,
      }),
    }),
  }),
  overrideExisting: false,
});
export { injectedRtkApi as petApi };
export type UploadFileApiResponse =
  /** status 200 successful operation */ ApiResponse;
export type UploadFileApiArg = {
  /** ID of pet to update */
  petId: number;
  body: {
    additionalMetadata?: string;
    file?: Blob;
  };
};
export type AddPetApiResponse = unknown;
export type AddPetApiArg = {
  /** Pet object that needs to be added to the store */
  pet: Pet;
};
export type UpdatePetApiResponse = unknown;
export type UpdatePetApiArg = {
  /** Pet object that needs to be added to the store */
  pet: Pet;
};
export type FindPetsByStatusApiResponse =
  /** status 200 successful operation */ Pet[];
export type FindPetsByStatusApiArg = {
  /** Status values that need to be considered for filter */
  status: ("available" | "pending" | "sold")[];
};
export type FindPetsByTagsApiResponse =
  /** status 200 successful operation */ Pet[];
export type FindPetsByTagsApiArg = {
  /** Tags to filter by */
  tags: string[];
};
export type GetPetByIdApiResponse = /** status 200 successful operation */ Pet;
export type GetPetByIdApiArg = {
  /** ID of pet to return */
  petId: number;
};
export type UpdatePetWithFormApiResponse = unknown;
export type UpdatePetWithFormApiArg = {
  /** ID of pet that needs to be updated */
  petId: number;
  body: {
    name?: string;
    status?: string;
  };
};
export type DeletePetApiResponse = unknown;
export type DeletePetApiArg = {
  apiKey?: string;
  /** Pet id to delete */
  petId: number;
};
export type PlaceOrderApiResponse =
  /** status 200 successful operation */ Order;
export type PlaceOrderApiArg = {
  /** order placed for purchasing the pet */
  order: Order;
};
export type GetOrderByIdApiResponse =
  /** status 200 successful operation */ Order;
export type GetOrderByIdApiArg = {
  /** ID of pet that needs to be fetched */
  orderId: number;
};
export type DeleteOrderApiResponse = unknown;
export type DeleteOrderApiArg = {
  /** ID of the order that needs to be deleted */
  orderId: number;
};
export type GetInventoryApiResponse = /** status 200 successful operation */ {
  [key: string]: number;
};
export type GetInventoryApiArg = void;
export type CreateUsersWithArrayInputApiResponse = unknown;
export type CreateUsersWithArrayInputApiArg = {
  /** List of user object */
  body: User[];
};
export type CreateUsersWithListInputApiResponse = unknown;
export type CreateUsersWithListInputApiArg = {
  /** List of user object */
  body: User[];
};
export type GetUserByNameApiResponse =
  /** status 200 successful operation */ User;
export type GetUserByNameApiArg = {
  /** The name that needs to be fetched. Use user1 for testing.  */
  username: string;
};
export type UpdateUserApiResponse = unknown;
export type UpdateUserApiArg = {
  /** name that need to be updated */
  username: string;
  /** Updated user object */
  user: User;
};
export type DeleteUserApiResponse = unknown;
export type DeleteUserApiArg = {
  /** The name that needs to be deleted */
  username: string;
};
export type LoginUserApiResponse =
  /** status 200 successful operation */ string;
export type LoginUserApiArg = {
  /** The user name for login */
  username: string;
  /** The password for login in clear text */
  password: string;
};
export type LogoutUserApiResponse = unknown;
export type LogoutUserApiArg = void;
export type CreateUserApiResponse = unknown;
export type CreateUserApiArg = {
  /** Created user object */
  user: User;
};
export type ApiResponse = {
  code?: number;
  type?: string;
  message?: string;
};
export type Category = {
  id?: number;
  name?: string;
};
export type Tag = {
  id?: number;
  name?: string;
};
export type Pet = {
  id?: number;
  category?: Category;
  name: string;
  photoUrls: string[];
  tags?: Tag[];
  status?: "available" | "pending" | "sold";
};
export type Order = {
  id?: number;
  petId?: number;
  quantity?: number;
  shipDate?: string;
  status?: "placed" | "approved" | "delivered";
  complete?: boolean;
};
export type User = {
  id?: number;
  username?: string;
  firstName?: string;
  lastName?: string;
  email?: string;
  password?: string;
  phone?: string;
  userStatus?: number;
};
export const {
  useUploadFileMutation,
  useAddPetMutation,
  useUpdatePetMutation,
  useFindPetsByStatusQuery,
  useFindPetsByTagsQuery,
  useGetPetByIdQuery,
  useUpdatePetWithFormMutation,
  useDeletePetMutation,
  usePlaceOrderMutation,
  useGetOrderByIdQuery,
  useDeleteOrderMutation,
  useGetInventoryQuery,
  useCreateUsersWithArrayInputMutation,
  useCreateUsersWithListInputMutation,
  useGetUserByNameQuery,
  useUpdateUserMutation,
  useDeleteUserMutation,
  useLoginUserQuery,
  useLogoutUserQuery,
  useCreateUserMutation,
} = injectedRtkApi;
