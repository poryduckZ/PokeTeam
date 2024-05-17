import ky from "ky";

const BASEURL = "http://localhost:8080/";

export const api = ky.create({
    prefixUrl: BASEURL,
});
