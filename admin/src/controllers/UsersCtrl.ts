import { UserSrvc, IUserItem, IUserResource } from "../services/UserSrvc";
import 'angular-resource';

export class UsersCtrl {

    private User: IUserResource;
    private users: IUserItem[] = [];

    static $inject = [
        'Users'
    ];
    constructor(
        UsersSrvc: UserSrvc
    ) {
        this.User = UsersSrvc.resource;
        this.users = this.User.query();
    }

}