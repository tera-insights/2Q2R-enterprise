import { Auth } from '../services/Auth';

export class LoginCtrl {

    login(userid: string, password: string) {



    }

    static $inject = ['Auth'];
    constructor(private Auth: Auth) {

    }

}