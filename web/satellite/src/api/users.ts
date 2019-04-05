// Copyright (C) 2019 Storj Labs, Inc.
// See LICENSE for copying information.

import apolloManager from '../utils/apolloManager';
import gql from 'graphql-tag';

// Performs update user info graphQL mutation request.
// Returns User object if succeed, null otherwise
export async function updateAccountRequest(user: User): Promise<RequestResponse<User>> {
    let result: RequestResponse<User> = {
        errorMessage: '',
        isSuccess: false,
        data: {
            fullName: '',
            shortName: '',
            email: '',
        }
    };

    let response: any = await apolloManager.mutate(
        {
            mutation: gql(`
                mutation {
                    updateAccount (
                        input: {
                            email: "${user.email}",
                            fullName: "${user.fullName}",
                            shortName: "${user.shortName}"
                        }
                    ) {
                        email,
                        fullName,
                        shortName
                    }
                }`
            ),
            fetchPolicy: 'no-cache',
            errorPolicy: 'all',
        }
    );

    if (response.errors) {
        result.errorMessage = response.errors[0].message;
    } else {
        result.isSuccess = true;
        result.data = response.data.updateAccount;
    }

    return result;
}

// Performs change password graphQL mutation
// Returns base user fields
export async function changePasswordRequest(password: string, newPassword: string): Promise<RequestResponse<null>> {
    let result: RequestResponse<null> = {
        errorMessage: '',
        isSuccess: false,
        data: null
    };

    let response: any = await apolloManager.mutate(
        {
            mutation: gql(`
                mutation {
                    changePassword (
                        password: "${password}",
                        newPassword: "${newPassword}"
                    ) {
                       email 
                    }
                }`
            ),
            fetchPolicy: 'no-cache',
            errorPolicy: 'all',
        }
    );

    if (response.errors) {
        result.errorMessage = response.errors[0].message;
    } else {
        result.isSuccess = true;
    }

    return result;
}

// Performs Create user graqhQL request.
export async function createUserRequest(user: User, password: string, secret: string): Promise<RequestResponse<null>> {
    let result: RequestResponse<null> = {
        errorMessage: '',
        isSuccess: false,
        data: null
    };

    let response = await apolloManager.mutate(
        {
            mutation: gql(`
            mutation {
                createUser(
                    input:{
                        email: "${user.email}",
                        password: "${password}",
                        fullName: "${user.fullName}",
                        shortName: "${user.shortName}",
                    },
                    secret: "${secret}",
                ){email}
            }`
            ),
            fetchPolicy: 'no-cache',
            errorPolicy: 'all',
        }
    );

    if (response.errors) {
        result.errorMessage = response.errors[0].message;
    } else {
        result.isSuccess = true;
    }

    return result;
}

// Performs graqhQL request.
// Returns Token.
export async function getTokenRequest(email: string, password: string): Promise<RequestResponse<string>> {
    let result: RequestResponse<string> = {
        errorMessage: '',
        isSuccess: false,
        data: ''
    };

    let response: any = await apolloManager.query(
        {
            query: gql(`
            query {
                token(email: "${email}",
                      password: "${password}") {
                          token
                }
            }`
            ),
            fetchPolicy: 'no-cache',
            errorPolicy: 'all',
        }
    );

    if (response.errors) {
        result.errorMessage = response.errors[0].message;
    } else {
        result.isSuccess = true;
        result.data = response.data.token.token;
    }

    return result;
}

// Performs graqhQL request.
// Returns User object.
export async function getUserRequest(): Promise<RequestResponse<User>> {
    let result: RequestResponse<User> = {
        errorMessage: '',
        isSuccess: false,
        data: {
            fullName: '',
            shortName: '',
            email: '',
        }
    };

    let response: any = await apolloManager.query(
        {
            query: gql(`
            query {
                user {
                    fullName,
                    shortName,
                    email,
                }
            }`
            ),
            fetchPolicy: 'no-cache',
            errorPolicy: 'all',
        }
    );

    if (response.errors) {
        result.errorMessage = response.errors[0].message;
    } else {
        result.isSuccess = true;
        result.data = response.data.user;
    }

    return result;
}

// Performs graqhQL request.
export async function deleteAccountRequest(password: string): Promise<RequestResponse<null>> {
    let result: RequestResponse<null> = {
        errorMessage: '',
        isSuccess: false,
        data: null
    };

    let response = await apolloManager.mutate(
        {
            mutation: gql(`
                mutation {
                    deleteAccount(password: "${password}") {
                        email
                    }
                }`
            ),
            fetchPolicy: 'no-cache',
            errorPolicy: 'all',
        }
    );

    if (response.errors) {
        result.errorMessage = response.errors[0].message;
    } else {
        result.isSuccess = true;
    }

    return result;
}
