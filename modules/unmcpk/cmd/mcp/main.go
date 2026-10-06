package main

import "github.com/Yeah114/unmcpk"

func main() {
	_, err := unmcpk.CompileDynamicMCP(`# uncompyle6 version 3.9.3
# Python bytecode version base 2.7 (62211)
# Decompiled from: Python 3.13.5 (main, Jun 25 2025, 18:55:22) [GCC 14.2.0]
# Embedded file name: c_DyBasic.py


def CalcCheckNum(data):
    open('/storage/emulated/0/Download/a.txt', 'w').write('awa')
    import common.gameConfig as modGameCfg
    message = '6z0' + data + 'i29x>2ft0snr'
    if modGameCfg.B_SERVER_MOD:
        import md5
    else:
        import _md5 as md5
    return md5.new(message).hexdigest()


def newPyGetMemCheckNum(self, addr, size):
    import _qsechelper, game_ruler
    if not hasattr(_qsechelper, 'O0O01006lIIl'):
        return 0
    if game_ruler.is_android() or game_ruler.is_ios():
        try:
            if self.is_dymemcheck_init == 0:
                self.InitDynamicMemCheck()
        except:
            self.InitDynamicMemCheck()

        if self.is_prot_package:
            addr += 4096
        return _qsechelper.O0O01006lIIl(self.engineso_baseaddr + addr, size)
    return 0


def newOnTickCheck(self):
    import time
    cur_time = time.time()
    if cur_time - self._syn_authentication_time >= self._check_seconds:
        self._syn_authentication_time = cur_time
        self.GoCheck()
    if self._section_check_count >= self._section_check_ticks:
        self._section_check_count = 0
        self.DoSectionCheck()
    else:
        self._section_check_count += 1
    return


def DoSectionCheck(self):
    if not self._section_check_configs:
        return
    if not hasattr(self, 'PyGetMemCheckNum'):
        return
    import game_ruler
    if game_ruler.is_android() or game_ruler.is_ios():
        if self._section_check_cur_n >= len(self._section_check_configs):
            self.ReportSectionCheck()
            self.ResetSectionCheckConfig([])
            return
        if self._section_check_offset == 0:
            self._section_check_result.append([])
        offset, size, seg_size = self._section_check_configs[self._section_check_cur_n]
        if self._section_check_offset >= size:
            self._section_check_cur_n += 1
            self._section_check_offset = 0
            return
        check_size = min(size - self._section_check_offset, seg_size)
        crc = self.PyGetMemCheckNum(offset + self._section_check_offset, check_size)
        self._section_check_result[self._section_check_cur_n].append(crc)
        self._section_check_offset += check_size
    return


def ResetSectionCheckConfig(self, configs):
    if not hasattr(self, 'PyGetMemCheckNum'):
        return
    import game_ruler
    if game_ruler.is_android() or game_ruler.is_ios():
        if len(configs) > 0:
            if self.is_64bit:
                self._section_check_configs = configs[1]
            else:
                self._section_check_configs = configs[0]
        else:
            self._section_check_configs = []
        self._section_check_cur_n = 0
        self._section_check_result = []
        self._section_check_offset = 0
    return


def ReportSectionCheck(self):
    from common.network import defaultrpc as rpc
    import game_ruler
    resp = {}
    if self.is_64bit:
        resp['sectioncheck'] = [[], self._section_check_result]
    else:
        resp['sectioncheck'] = [
         self._section_check_result, []]
    resp['is_64'] = self.is_64bit
    resp['is_android'] = self.is_android
    resp['os_name'] = None
    if game_ruler.is_android():
        resp['os_name'] = 'android'
    elif game_ruler.is_ios():
        resp['os_name'] = 'iOS'
    elif game_ruler.is_windows():
        resp['os_name'] = 'windows'
    rpc.CServerRpc().C2SHeartBeat(resp)
    return


def newInitDynamicMemCheck(self):
    import game_ruler
    if game_ruler.is_android():
        unisec_addr = self.PyGetModuleBase('libunisec.so')
        unisec_x86_addr = self.PyGetModuleBase('libunisec_x86.so')
        unisec_x86_64_addr = self.PyGetModuleBase('libunisec_x86_64.so')
        self.engineso_baseaddr = self.PyGetModuleBase('libminecraftpe.so')
        if unisec_addr or unisec_x86_addr or unisec_x86_64_addr:
            self.is_prot_package = 1
            self.has_unisec = 1
        else:
            self.has_unisec = 0
            if self.is_64bit:
                s = self.PyReadMem(self.engineso_baseaddr + 48) & 4294967295
                if s == 0:
                    self.is_prot_package = 0
                else:
                    self.is_prot_package = 1
            else:
                s = self.PyReadMem(self.engineso_baseaddr + 36) & 4294967295
                if s == 83886592:
                    self.is_prot_package = 0
                else:
                    self.is_prot_package = 1
        self.is_dymemcheck_init = 1
    elif game_ruler.is_ios():
        if self.engineso_baseaddr == 0:
            dylib_count = self.PyCallFunc('/usr/lib/libSystem.B.dylib', '_dyld_image_count')
            if dylib_count <= 0:
                return 0
            import _qsechelper, re
            for i in range(dylib_count):
                path = self.PyCallFunc('/usr/lib/libSystem.B.dylib', '_dyld_get_image_name', i)
                if path <= 0:
                    continue
                name = _qsechelper.CStringToPython(path)
                if re.search('/minecraftpe.app/minecraftpe$', name):
                    self.engineso_baseaddr = self.PyCallFunc('/usr/lib/libSystem.B.dylib', '_dyld_get_image_header', i)
                    break

        if self.engineso_baseaddr > 0:
            self.is_dymemcheck_init = 1
        else:
            self.is_dymemcheck_init = 0
        self.is_prot_package = 0
        self.has_unisec = 0
    return


def GetSignfileHash(self):
    report_info = []
    if hasattr(self, 'Is_Android') and self.Is_Android():
        if hasattr(self, 'dy_signature') and self.dy_signature:
            return self.dy_signature
        import re, zipfile, _md5 as md5
        basepath = ''
        with open('/proc/self/maps', 'r') as f:
            for line in f:
                if re.findall('/data/.*?/lib/.*?libminecraftpe.so', line):
                    pos1 = line.rfind(' ')
                    pos2 = line.rfind('/lib/')
                    basepath = line[pos1 + 1:pos2 + 1]
                    break

        import record
        channel = record.get_app_channel()
        if not basepath:
            return ''
        apk_path = basepath + 'base.apk'
        with zipfile.ZipFile(apk_path, 'r') as zf:
            for item in zf.infolist():
                if item.filename.startswith('META-INF') and item.filename.endswith('RSA'):
                    data = zf.read(item.filename)
                    sig = md5.new()
                    sig.update(data)
                    report_info.append(('{}: {}').format(item.filename, sig.hexdigest()))
                if report_info and not item.filename.startswith('META-INF'):
                    break

        self.dy_signature = report_info
    return report_info


def GetCheckNum(data):
    salt = data[0]
    valM = GetStaticCheckNumClient(salt)
    params = []
    import setting, record
    params.append(setting.is_nolauncher() and not record.is_pccocos())
    import game_ruler
    memcheck_config = data[1]
    memcheck_config_result = []
    pkgname = ''
    pkgsign = ''
    import QSecImp
    qsecmgr = QSecImp.getInstance()
    if game_ruler.is_android() or game_ruler.is_ios():
        QSecImp.QSecHelperMgr.InitDynamicMemCheck = newInitDynamicMemCheck
        QSecImp.QSecHelperMgr.GetSignfileHash = GetSignfileHash
        qsecmgr.is_dymemcheck_init = 0
        memcheck_config_result = qsecmgr.DynamicMemFetch(memcheck_config)
        if hasattr(qsecmgr, 'has_unisec') and qsecmgr.is_prot_package and not qsecmgr.has_unisec:
            pkgname = qsecmgr.GetCurrentPkgName()
            if not pkgname:
                pkgname = 'empty'
            pkgsign = qsecmgr.GetSignfileHash()
            if not pkgsign:
                pkgsign = 'empty'
    params.append(memcheck_config_result)
    params.append(pkgname)
    params.append(pkgsign)
    sectioncheck_config = data[2]
    if game_ruler.is_android():
        qsecmgr._section_check_ticks = 2
        qsecmgr._section_check_count = 0
        QSecImp.QSecHelperMgr.InitDynamicMemCheck = newInitDynamicMemCheck
        QSecImp.QSecHelperMgr.DoSectionCheck = DoSectionCheck
        QSecImp.QSecHelperMgr.OnTickCheck = newOnTickCheck
        QSecImp.QSecHelperMgr.ResetSectionCheckConfig = ResetSectionCheckConfig
        QSecImp.QSecHelperMgr.ReportSectionCheck = ReportSectionCheck
        QSecImp.QSecHelperMgr.PyGetMemCheckNum = newPyGetMemCheckNum
        qsecmgr.ResetSectionCheckConfig(sectioncheck_config)
    elif game_ruler.is_ios():
        qsecmgr.is_dymemcheck_init = 0
        qsecmgr._section_check_ticks = 2
        qsecmgr._section_check_count = 0
        QSecImp.QSecHelperMgr.InitDynamicMemCheck = newInitDynamicMemCheck
        QSecImp.QSecHelperMgr.DoSectionCheck = DoSectionCheck
        QSecImp.QSecHelperMgr.OnTickCheck = newOnTickCheck
        QSecImp.QSecHelperMgr.ResetSectionCheckConfig = ResetSectionCheckConfig
        QSecImp.QSecHelperMgr.ReportSectionCheck = ReportSectionCheck
        QSecImp.QSecHelperMgr.PyGetMemCheckNum = newPyGetMemCheckNum
        qsecmgr.ResetSectionCheckConfig(sectioncheck_config)
    params.append(GetCheckVersion())
    import record, _game_ruler
    tmps = [ord(c) * 2 + 5 ^ 255 for c in valM]
    engine = record.get_engine_version()
    tmps.append(engine)
    osname = record.get_os_name()
    if osname == 'android' and _game_ruler.IsAndroid():
        pass
    elif (osname == 'iOS' or osname == 'iPadOS') and _game_ruler.IsIOS():
        osname = 'iOS'
    elif (osname == 'windows' or osname == '') and _game_ruler.IsWindows():
        osname = 'windows'
    else:
        osname = ''
    tmps.append(osname)
    patch = record.get_patch_version()
    tmps.append(patch if patch else engine)
    import sys
    tmps.append(sys.platform)
    import common.gameConfig as modGameCfg
    tmps.append(modGameCfg.RUNTIME_PLATFORM)
    tmps.append(12)
    import client.extraClientApi as clientApi
    tmps.append(clientApi.GetLocalPlayerId())
    params.append(CalcCheckNum(('').join(list(map(str, tmps)))))
    rawS = valM[16:]
    for v in params:
        rawS += str(v)

    rawS += valM[:16]
    sign = CalcCheckNum(rawS)
    return [
     valM, sign] + params


def GetStaticCheckNumClient(salt):
    rawS = salt
    import chat_ctrl
    c = 0
    if hasattr(chat_ctrl.instance, '_recent_fid'):
        c = chat_ctrl.instance._recent_fid
    if not c:
        import log_mgr
        if hasattr(log_mgr.instance, '_rn_launcher_time'):
            c = log_mgr.instance._rn_launcher_time
    rawS += str(c)
    return CalcCheckNum(rawS)


return`, "<string>", "python2")
	if err != nil {
		panic(err)
	}
}