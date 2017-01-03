/// <reference path="../../typings/index.d.ts" />

import { Users, IUserItem, IUserResource } from "../services/Users";

export class UsersCtrl {

    private User: IUserResource;
    private users: IUserItem[] = [];

    static $inject = [
        'Users'
    ];
    constructor(
        UsersSrvc: Users
    ) {
        this.User = UsersSrvc.resource;
        this.users = this.User.query();
    }

}