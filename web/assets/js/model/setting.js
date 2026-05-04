class AllSetting {

    constructor(data) {
        this.pageSize = 25;
        this.expireDiff = 0;
        this.trafficDiff = 0;
        this.remarkModel = "-ieo";
        this.twoFactorEnable = false;
        this.twoFactorToken = "";
        this.xrayTemplateConfig = "";
        this.subCustomHeaders = "";
        this.subCustomHtml = "";
        this.subCustomErrorHtml = "";
        this.subEnableIndexPage = false;
        this.subIndexPageHtml = "";
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