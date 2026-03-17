class AllSetting {

    constructor(data) {
        this.webListen = "";
        this.webDomain = "";
        this.webPort = 2053;
        this.webCertFile = "";
        this.webKeyFile = "";
        this.webBasePath = "/";
        this.sessionMaxAge = 360;
        this.pageSize = 25;
        this.expireDiff = 0;
        this.trafficDiff = 0;
        this.remarkModel = "-ieo";
        this.twoFactorEnable = false;
        this.twoFactorToken = "";
        this.xrayTemplateConfig = "";
        this.subEnable = true;
        this.subTitle = "";
        this.subCustomHeaders = "";
        this.subCustomHtml = "";
        this.subListen = "";
        this.subPort = 2096;
        this.subPath = "/sub/";
        this.subDomain = "";
        this.externalTrafficInformEnable = false;
        this.externalTrafficInformURI = "";
        this.subCertFile = "";
        this.subKeyFile = "";
        this.subUpdates = 12;
        this.subEncrypt = true;
        this.subShowInfo = true;
        this.subURI = "";

        this.timeLocation = "Local";

        if (data == null) {
            return
        }
        ObjectUtil.cloneProps(this, data);
    }

    equals(other) {
        return ObjectUtil.equals(this, other);
    }
}