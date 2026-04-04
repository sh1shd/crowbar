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
        this.subCustomHeaders = "";
        this.subCustomHtml = "";
        this.subCustomErrorHtml = "";
        this.subEnableIndexPage = false;
        this.subIndexPageHtml = "";
        this.subListen = "";
        this.subPort = 2096;
        this.subPath = "/sub/";
        this.subDomain = "";
        this.externalTrafficInformEnable = false;
        this.externalTrafficInformURI = "";
        this.subCertFile = "";
        this.subKeyFile = "";
        this.subEncrypt = true;
        this.subURI = "";
        this.subMessageClientDisabled = "Disabled";
        this.subMessageClientExpired = "Expired";
        this.subMessageClientTrafficEnd = "Traffic has ended";
        this.subMessageContactAdmin = "Please contact administrator";

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