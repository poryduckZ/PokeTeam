import ky from "ky";

// TODO: Change dynamically based on environment
const BASEURL = "http://localhost:8080/";

export const api = ky.create({
    prefixUrl: BASEURL,
});
