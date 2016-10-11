/// <reference path="../../typings/index.d.ts" />
/// <reference path="../services/Users.ts" />

module admin {

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

}