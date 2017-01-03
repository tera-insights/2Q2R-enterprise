import { Users, IUserItem, IUserResource } from "../services/Users";
import 'angular-resource';

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