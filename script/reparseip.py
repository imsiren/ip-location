#-*- encoding=utf8
#每天定时下载ip数据库，lbs服务会自动监测文件修改，并重新加载ip库以实现天维度的数据更新
#该文件下载完成会校验token，只有token匹配才视为有效文件
# TODO :增加下载失败邮件监控

import requests
import time
import hashlib,sys,os,shutil,struct

from util import bytes2long, ip2long, convert, verify_ipv4,long2bytes

#if sys.getdefaultencoding() != 'utf-8':
reload(sys)
sys.setdefaultencoding('utf-8')

class RestoryIpData:

    data = b""
    indexsize = 0
    areaCodes = {}
    newSize=[]

    def __init__(self, ipname, codename):

        self.loadIpDatx(ipname)   

        self.loadCityCode(codename)

        self.restoreIpDatx()

    def loadIpDatx(self, ipname):
        #解析索引数据
        file = open(ipname, "rb")
        self.data = file.read()
        file.close()

        #print ord(self.data[0]), ord(self.data[1]), ord(self.data[2]), ord(self.data[3])
        self.indexSize = bytes2long((self.data[0]), (self.data[1]), (self.data[2]), (self.data[3]))
        #print self.indexSize
        #sys.exit(0)

   #重建datx数据结构 
    def restoreIpDatx(self):

        fp = open("data/test.datx","wb")

        a, b, c, d = long2bytes(self.indexSize)
        fp.write(a)
        fp.write(b)
        fp.write(c)
        fp.write(d)
        fp.write(self.data[0:262148])
        #fp.close()
        #self.loadIpDatx("./data/test.datx")

        #sys.exit(0)
        
        high = int((self.indexSize - 262144 - 262148) / 9) - 1
        #fp.write(self.data[26219:])
        #print self.data[26203559+10000:26203559+11000]
        writedata= ""
        for i in range(high):
            pos = i * 9 + 262148
            off = convert(self.data[pos+6]) << 16 | convert(self.data[pos+5]) << 8 | convert(self.data[pos+4])
            l = convert(self.data[pos+7]) << 8 | convert(self.data[pos+8])

            pos = off - 262144 + self.indexSize
            
            print pos, off, l,self.indexSize
            sys.exit(0)
            tmp = (self.data[pos:pos+l])
            #if len(tmp) == 0:
            #    continue
            #.decode("utf-8")
            #area = tmp.split("\t")
            #print pos, pos+l, off
            #sys.exit(0)
            #cityCode = self.findCityCode(area[0], area[1], area[2])
            #if len(cityCode) == 0:
            #    continue
            #area[9] = cityCode[0]
            #area[14] = cityCode[9]
            #area.insert(14, cityCode[9])
            #fp.write("\t".join(area))
            #fp.write(tmp)
            writedata += tmp
            #print "\t".join(tmp)
        fp.write(struct.pack("@8s",writedata))
        fp.close()


    #解析city.txt，加载地区code
    def loadCityCode(self,codename):

        fp = open(codename)

        line = fp.readline()
        while line:
            array = line.split("\t")
            country = array[1]
            self.areaCodes[country] = array
            line = fp.readline()
        fp.close()

    #查找城市码
    def findCityCode(self, country, province, city):
        #国家必须存在
        if country == "":
            return []
        if country in self.areaCodes:
            data = self.areaCodes[country]
        else:
            return []

        if len(data) == 0:
            return []
        index = 1
        if city != "":
            #城市
            index = 7
            name = city
        if city == "":
            #省
            index = 4
            name = province
        if province == "":
            #国家
            index = 1
            name = country
        for i in range(len(data)):
            if data[index] == name:
                return data
        return []

    def downloadIpDatx(self):
        down_url = "" # 下载地址
        response = requests.get(down_url, verify=False) # 发起http请求
        etag_value = response.headers.get("ETag") #获取ETag值
        content_length = int(response.headers.get("Content-Length"))
        current_length = 0
        read_size = 4096
        if not etag_value:           # ETag不存在就退出
            print("etag not exists")
            sys.exit(0)
        with open("data/ipip_temp.datx", 'wb+') as fd: #写临时文件
            for chunk in response.iter_content(4096):
                fd.write(chunk)
        with open("data/ipip_temp.datx", 'rb') as fd: #读取临时文件
            sha1 = hashlib.sha1()
            while True:
                content = fd.read(read_size)
                if not content:
                    break
                sha1.update(content)
                current_length += read_size
                sys.stdout.write("%.2f" % int(float(current_length)/float(content_length) * 100.0)+ "%\r")
                sys.stdout.flush()
                time.sleep(0.01)
            content_sha1_value = sha1.hexdigest() #计算临时文件sha1
            etag_sha1_value = etag_value[5:]
            if etag_sha1_value != content_sha1_value: # sha1 不一致退出
                print("etag err , please try again")
                sys.exit(0)
        shutil.move("data/ipip_temp.datx", "data/ipip_station.datx") # 覆盖正式文件，目标目录必须有可写权限。
        print("ok")

Story = RestoryIpData("./data/ipip_station.datx", "./data/city.txt")
#Story = RestoryIpData("./data/test.datx", "./data/city.txt")


